package probe

import (
	"fmt"
	"strings"
	"time"

	"github.com/csduarte/mattermost-probe/config"
	"github.com/csduarte/mattermost-probe/mattermost"
	"github.com/csduarte/mattermost-probe/metrics"
	"github.com/csduarte/mattermost-probe/util"
	"github.com/mattermost/platform/model"
	uuid "github.com/satori/go.uuid"
)

// BroadcastProbe represents a test where the speaker will broadcast unique messages and the listener will check broadcast time.
type BroadcastProbe struct {
	Speaker       *mattermost.Client
	Listener      *mattermost.Client
	Config        *config.BroadcastConfig
	Messages      *util.MessageMap
	EventChannel  chan *model.WebSocketEvent
	TimingChannel metrics.TimingChannel
	StopChannel   chan bool
	Active        bool
}

// NewBroadcastProbe creates a new base probe
func NewBroadcastProbe(c *config.BroadcastConfig, s, l *mattermost.Client) *BroadcastProbe {
	bp := &BroadcastProbe{
		s,
		l,
		c,
		util.NewMessageMap(),
		make(chan *model.WebSocketEvent, 10),
		nil,
		make(chan bool),
		false,
	}

	return bp
}

// Setup will run once on application starts
func (bp *BroadcastProbe) Setup() error {
	if len(bp.Config.ChannelID) < 1 && len(bp.Config.ChannelName) < 1 {
		return fmt.Errorf("Must set either ChannelID or ChannelName for probe")
	}

	if len(bp.Config.ChannelID) < 1 {
		err := bp.getChannelID(bp.Config.ChannelName)
		if err != nil {
			return err
		}
	}

	if err := bp.ensureMembership(bp.Listener); err != nil {
		return err
	}
	if err := bp.ensureMembership(bp.Speaker); err != nil {
		return err
	}

	bp.Listener.AddSubscription(bp)

	return nil
}

// Start will kick off the probe
func (bp *BroadcastProbe) Start() error {

	if bp.Active {
		return nil
	}

	go bp.listenForEvents()

	writeTicker := time.NewTicker(time.Millisecond * bp.Config.Frequency)
	go func() {
		for {
			select {
			case <-bp.StopChannel:
				return
			case <-writeTicker.C:
				go bp.SendWrite()
			}
		}
	}()

	bp.Active = true
	return nil
}

// SendWrite sends a sample post
func (bp *BroadcastProbe) SendWrite() {
	p := &model.Post{}
	uid := uuid.NewV4().String()
	sentAt := time.Now()
	bp.Messages.Add(uid, sentAt)
	p.ChannelId = bp.Config.ChannelID
	p.UserId = bp.Speaker.User.Id
	p.Message = uid
	if err := bp.Speaker.CreatePost(p); err != nil {
		bp.Speaker.LogError("Error while SendWrite", err.Error())
	}
}

func (bp *BroadcastProbe) listenForEvents() {
	for {
		select {
		case e := <-bp.EventChannel:
			bp.handleEvent(e)
		}
	}
}

func (bp *BroadcastProbe) handleEvent(event *model.WebSocketEvent) {
	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	uid := post.Message
	end := time.Now()
	start, ok := bp.Messages.Delete(uid)
	if !ok {
		bp.Speaker.LogError("Failed to find message by uuid")
	}
	if bp.TimingChannel != nil {
		bp.TimingChannel <- metrics.TimingReport{
			MetricName:      metrics.MetricProbeBroadcast,
			DurationSeconds: end.Sub(start).Seconds(),
		}
	}
}

func (bp *BroadcastProbe) getChannelID(name string) error {
	channel, err := bp.Speaker.GetChannelByName(name)
	if err != nil {
		bp.Speaker.LogError("Probe error", err.Error())
	}

	bp.Config.ChannelID = channel.Id
	return nil
}

func (bp *BroadcastProbe) ensureMembership(c *mattermost.Client) error {
	err := c.JoinChannel(bp.Config.ChannelID)
	if err != nil {
		return err
	}
	return err
}

// GetSubscription adheres to SubscriptionProbe from mattermost subpackag
func (bp BroadcastProbe) GetSubscription() *mattermost.WebSocketSubscription {
	wss := mattermost.NewWebsocketSubcription(bp.EventChannel)
	// TODO: Create append helper functions
	wss.UserIDs = append(wss.UserIDs, bp.Speaker.User.Id)
	wss.ChannelIDs = append(wss.ChannelIDs, bp.Config.ChannelID)
	wss.EventTypes = append(wss.EventTypes, model.WEBSOCKET_EVENT_POSTED)
	return wss
}

// func (wc *WriteCheck) Stop() {
// 	wc.StopChannel <- true
// }
// CheckOverdue will handle any overdue messages
// func (bp *BroadcastProbe) CheckOverdue() {
// 	for {
// 		if id, delay := bp.Messages.FistOverdue(50); delay > 0 {
// 			bp.Messages.Delete(id)
// 			fmt.Printf("WARN: SLOW MESAGE took %v ms", delay)
// 			continue
// 		}
// 		break
// 	}
// }

// from start
// overdueTicker := time.NewTicker(time.Millisecond * 500)
// go func() {
// 	for {
// 		select {
// 		case <-bp.StopChannel:
// 			return
// 		case <-overdueTicker.C:
// 			go bp.CheckOverdue()
// 		}
// 	}
// }()
