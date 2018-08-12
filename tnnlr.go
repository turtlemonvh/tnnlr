package tnnlr

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/gin-gonic/gin.v1"
	"labix.org/v2/mgo/bson"
)

// Store messages for the user in between views
// FIXME: Add fields to improve UI
type Message struct {
	msg string
	mtime time.Time
}

func (m *Message) Mstring() string {
	return m.msg
}

func (m *Message) Tstring() string {
	return m.mtime.Format(time.RFC822)
}


// Server
type Tnnlr struct {
	sync.Mutex
	Template         *template.Template
	SshExec          string // path to ssh executable
	LogLevel         string
	TunnelReloadFile string
	Port int
	msgs             chan Message
	tunnels          map[string]*Tunnel
}

func (t *Tnnlr) Init() {
	var err error
	var level log.Level

	// Create bookkeeping directories
	for _, dir := range []string{relProc, relLog} {
		if err = createRelDir(dir); err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"dir": dir,
			}).Fatal("Unable to create bookkeeping directory")
		}
	}

	// Parse template
	t.Template, err = template.New("Homepage").Parse(homePage)
	if err != nil {
		log.Fatal(err)
	}

	// Set log level
	level, err = log.ParseLevel(t.LogLevel)
	if err != nil {
		log.WithFields(log.Fields{
			"err":          err,
			"levelOptions": log.AllLevels,
		}).Fatal("Invalid value for option 'log-level'")
	}
	log.SetLevel(level)

	// A generously buffered channel
	t.msgs = make(chan Message, 100)
	t.tunnels = make(map[string]*Tunnel)

	// FIXME: Pull from config
	// ADD: default username
	t.TunnelReloadFile = ".tnnlr"
	t.SshExec = "ssh"
}

// Add a message to the queue
// If the channel is full, the message is logged and discarded
func (t *Tnnlr) AddMessage(msg string) {
	select {
	case t.msgs <- Message{msg, time.Now()}:
		log.WithFields(log.Fields{
			"nmsgs": len(t.msgs),
			"msg": msg,
		}).Debug("Added message.")
	case <-time.After(1*time.Millisecond):
		log.WithFields(log.Fields{
			"nmsgs": len(t.msgs),
			"msg": msg,
		}).Error("Message buffer is full, can't add message. Reload page to drain messages.")
	}
}

func (t *Tnnlr) Run() {
	// Launch process to clean logs and pid files
	go t.CleanBookkeepingDirs()

	if log.GetLevel() != log.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.GET("/", t.HomepageView)
	r.POST("/save", t.Save)
	r.POST("/add", t.Add)
	r.GET("/remove/:id", t.Remove)
	r.POST("/reload", t.Reload)
	r.GET("/reload/:id", t.ReloadOne)
	r.GET("/bash_command/:id", t.ShowCommand)
	r.GET("/logs/:id", t.ShowLogs)
	r.GET("/status/:id", t.ReloadOne)
	r.Run(fmt.Sprintf(":%d", t.Port))
}

// HTTP views

func (t *Tnnlr) HomepageView(c *gin.Context) {
	var messages []Message
messageLoop:
	for {
		select {
		case msg := <-t.msgs:
			messages = append(messages, msg)
		case <-time.After(1 * time.Millisecond):
			break messageLoop
		}
	}

	data := struct {
		HasMessages bool
		Messages    []Message
		Tunnels     map[string]*Tunnel
	}{
		len(messages) > 0,
		messages,
		t.tunnels,
	}

	if err := t.Template.Execute(c.Writer, data); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Error executing template")
	}
}

// Reload from config file
func (t *Tnnlr) Reload(c *gin.Context) {
	log.Warn("Killing all active tunnels before loading")
	t.KillAllTunnels()

	tmpTunnels, err := t.Load()
	if err != nil {
		message := "Failed to parse tunnels from file"
		log.WithFields(log.Fields{
			"err":  err.Error(),
			"file": t.TunnelReloadFile,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Add 1 at a time
	nTunnelsLoaded := 0
	for _, tnnl := range tmpTunnels {
		if err = t.AddTunnel(tnnl); err != nil {
			message := fmt.Sprintf("Failed to add tunnel '%s' from file", tnnl.Id)
			log.WithFields(log.Fields{
				"err":  err.Error(),
				"file": t.TunnelReloadFile,
			}).Error(message)
			t.AddMessage(message)
		} else {
			nTunnelsLoaded++
		}
	}

	t.AddMessage(fmt.Sprintf("Finished loading %d of %d tunnels from file: %s", nTunnelsLoaded, len(tmpTunnels), t.TunnelReloadFile))
	c.Redirect(http.StatusFound, "/")
}

// Save set of tunnels to a file
func (t *Tnnlr) Save(c *gin.Context) {
	f, err := os.Create(t.TunnelReloadFile)
	if err != nil {
		message := "Failed to open tunnel file"
		log.WithFields(log.Fields{
			"err":  err.Error(),
			"file": t.TunnelReloadFile,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	var tmpTunnels []*Tunnel
	for _, tnnl := range t.tunnels {
		tmpTunnels = append(tmpTunnels, tnnl)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(tmpTunnels)
	if err != nil {
		message := "Failed to write json to file"
		log.WithFields(log.Fields{
			"err":  err.Error(),
			"file": t.TunnelReloadFile,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	t.AddMessage(fmt.Sprintf("Successfully saved tunnels to file: %s", t.TunnelReloadFile))
	c.Redirect(http.StatusFound, "/")
}

// Stop and remove a single tunnel
func (t *Tnnlr) Remove(c *gin.Context) {
	tnnlId := c.Param("id")

	// Error are handled in this function itself
	t.RemoveTunnel(tnnlId)
	c.Redirect(http.StatusFound, "/")
}

// Reload a single tunnel from disk
func (t *Tnnlr) ReloadOne(c *gin.Context) {
	rTnnlId := c.Param("id")

	tmpTunnels, err := t.Load()
	if err != nil {
		message := "Failed to parse tunnels from file"
		log.WithFields(log.Fields{
			"err":  err.Error(),
			"file": t.TunnelReloadFile,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Find the tunnel requested
	var foundTnnl Tunnel
	for _, tnnl := range tmpTunnels {
		if tnnl.Id == rTnnlId {
			foundTnnl = tnnl
			break
		}
	}
	if foundTnnl.Id == "" {
		message := "Failed to find tunnel with the requested id"
		log.WithFields(log.Fields{
			"file": t.TunnelReloadFile,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Errors are handled in the functions themselves
	// Stop process if it is running now
	t.RemoveTunnel(rTnnlId)
	t.AddTunnel(foundTnnl)

	c.Redirect(http.StatusFound, "/")
}

// Add a new tunnel
// Accepts html form or json
func (t *Tnnlr) Add(c *gin.Context) {
	var newTunnel Tunnel
	var err error

	err = c.Bind(&newTunnel)
	if err != nil {
		message := "Invalid form submission"
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Ids are bson by default
	err = t.AddTunnel(newTunnel)
	if err != nil {
		message := fmt.Sprintf("Unable to add tunnel: %s", newTunnel.Id)
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	t.AddMessage(fmt.Sprintf("Successfully added new tunnel: %s", newTunnel.Id))
	c.Redirect(http.StatusFound, "/")
}

func (t *Tnnlr) ShowCommand(c *gin.Context) {
	rTnnlId := c.Param("id")

	log.WithFields(log.Fields{
		"id":      rTnnlId,
		"tunnels": t.tunnels,
	}).Info("Showing command for tunnel")

	tnnl, ok := t.tunnels[rTnnlId]
	if !ok {
		message := "Failed to find tunnel with the requested id"
		log.WithFields(log.Fields{
			"file": t.TunnelReloadFile,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.String(200, tnnl.getCommand())
}

func (t *Tnnlr) ShowLogs(c *gin.Context) {
	rTnnlId := c.Param("id")

	log.WithFields(log.Fields{
		"id":      rTnnlId,
		"tunnels": t.tunnels,
	}).Info("Showing logs for tunnel")

	tnnl, ok := t.tunnels[rTnnlId]
	if !ok {
		message := "Failed to find tunnel with the requested id"
		log.WithFields(log.Fields{
			"file": t.TunnelReloadFile,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}

	logfilePath, err := tnnl.LogPath()
	if err != nil {
		message := "Failed to find logfile for tunnel with the requested id"
		log.WithFields(log.Fields{
			"Id": tnnl.Id,
		}).Error(message)
		t.AddMessage(message)
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.File(logfilePath)
}

// Load from disk
func (t *Tnnlr) Load() ([]Tunnel, error) {
	var tmpTunnels []Tunnel

	// Load from file
	raw, err := ioutil.ReadFile(t.TunnelReloadFile)
	if err != nil {
		return tmpTunnels, err
	}

	// Load all tunnels
	err = json.Unmarshal(raw, &tmpTunnels)

	return tmpTunnels, err
}

// Add a single tunnel
// Threadsafe
func (t *Tnnlr) AddTunnel(tnnl Tunnel) error {
	// Validate
	if err := tnnl.Validate(); err != nil {
		return err
	}

	// Add id if it doesn't have one
	if tnnl.Id == "" {
		tnnl.Id = bson.NewObjectId().Hex()
	}

	// Startup
	if err := tnnl.Run(t.SshExec); err != nil {
		return err
	}

	// Include in map
	t.Lock()
	t.tunnels[tnnl.Id] = &tnnl
	t.Unlock()

	return nil
}

// Remove a single tunnel
// Threadsafe
// Logs errors stopping the process, but continues
func (t *Tnnlr) RemoveTunnel(tnnlId string) error {
	var err error
	t.Lock()
	tnnl, ok := t.tunnels[tnnlId]
	if !ok {
		message := fmt.Sprintf("Did not find any tunnel with id: %s", tnnlId)
		t.AddMessage(message)
		return fmt.Errorf(message)
	}

	if err = tnnl.Stop(); err != nil {
		t.AddMessage(fmt.Sprintf("Failed to kill tunnel %s: '%s'", tnnl.Id, tnnl.Name))
	}
	delete(t.tunnels, tnnl.Id)
	t.Unlock()
	return err
}

// Kill all active tunnels
func (t *Tnnlr) KillAllTunnels() {
	t.Lock()
	for tnnlId, _ := range t.tunnels {
		t.RemoveTunnel(tnnlId)
	}
	t.Unlock()
}

func (t *Tnnlr) ManagedTunnels() map[string]bool {
	t.Lock()
	var tunnelIds = make(map[string]bool)
	for tnnlId, _ := range t.tunnels {
		tunnelIds[tnnlId] = true
	}
	t.Unlock()
	return tunnelIds
}

/*
Background cleanup and management of jobs

TODO
- check cmd and pid information directly to see if process is running instead of checking port
- option to leave process running when tnnlr shuts down

*/
func (t *Tnnlr) CleanBookkeepingDirs() {

	procDir, err := getRelativePath(relProc)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Unable to check proc dir to reap processes")
		return
	}

	logDir, err := getRelativePath(relLog)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Unable to check log dir to clean up logs")
		return
	}

	for {
		// Loop every 10 seconds
		time.Sleep(10 * time.Second)

		// Based on scanning pid dir
		runningProcesses := make(map[string]bool)
		// From the state of the program
		managedProcesses := t.ManagedTunnels()

		log.WithFields(log.Fields{
			"dir": procDir,
		}).Debug("Scanning pid dir for changed files")

		// Load all tunnels
		// Clean up any pid files associated with tunnels that are not currently running and are not known to the current live process
		pidFiles, err := filepath.Glob(fmt.Sprintf("%s/*.pid", procDir))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Error listing files in proc dir")
			continue
		}

		for _, pf := range pidFiles {
			c, e := ioutil.ReadFile(pf)
			if e != nil {
				os.Remove(pf)
			}
			var tnnl Tunnel
			err = json.Unmarshal(c, &tnnl)
			if e != nil {
				os.Remove(pf)
			}
			// Check if it this process is running
			if !tnnl.PortInUse() {
				// The process is dead.
				// Check if we should be running this and restart.
				if managedProcesses[tnnl.Id] {
					// Should restart
					log.WithFields(log.Fields{
						"id":   tnnl.Id,
						"name": tnnl.Name,
						"cmd":  tnnl.getCommand(),
					}).Info("Found dead process, restarting")
					tnnl.Run(t.SshExec)
					runningProcesses[tnnl.Id] = true
				} else {
					// Cleanup
					log.WithFields(log.Fields{
						"id":   tnnl.Id,
						"name": tnnl.Name,
						"cmd":  tnnl.getCommand(),
					}).Info("Found dead process, cleaning up")
					tnnl.Stop()
				}
			} else {
				log.WithFields(log.Fields{
					"id":   tnnl.Id,
					"name": tnnl.Name,
					"cmd":  tnnl.getCommand(),
				}).Debug("Found running process in pid dir")
				runningProcesses[tnnl.Id] = true
			}
		}

		// Load all logfiles
		// Clean up any not associated with a process that is live or being restarted
		logFiles, err := filepath.Glob(fmt.Sprintf("%s/*.log", logDir))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Error listing files in log dir")
			continue
		}

		for _, lf := range logFiles {
			// Get the id from the filename
			pts := strings.Split(lf, "/")
			lpt := pts[len(pts)-1]
			id := strings.Split(lpt, ".log")[0]

			// If this isn't running, delete it
			if !runningProcesses[id] {
				log.WithFields(log.Fields{
					"logFile": lf,
				}).Info("Removing logfile for dead process")
				os.Remove(lf)
			}
		}

	}
}