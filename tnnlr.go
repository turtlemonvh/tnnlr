package tnnlr

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
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
}

func (m *Message) String() string {
	return m.msg
}

// Server
type Tnnlr struct {
	sync.Mutex
	Template         *template.Template
	SshExec          string // path to ssh executable
	LogLevel         string
	TunnelReloadFile string
	msgs             chan Message
	tunnels          map[string]*Tunnel
}

func (t *Tnnlr) Init() {
	var err error
	var level log.Level

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
func (t *Tnnlr) AddMessage(msg string) {
	select {
	case t.msgs <- Message{msg}:
	case <-time.After(time.Second):
		log.WithFields(log.Fields{
			"nmsgs": len(t.msgs),
		}).Error("Message buffer is full. Reload page to drain messages.")
	}
}

func (t *Tnnlr) Run() {
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
	r.GET("/status/:id", t.ReloadOne)
	r.Run()
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

	c.JSON(200, gin.H{
		"cmd": tnnl.getCommand(),
	})
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
