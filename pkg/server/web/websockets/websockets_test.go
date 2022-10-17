package websockets

import (
	"sync"
	"testing"

	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestNotify(t *testing.T) {
	c := &channels{
		log:     zerolog.New().WithPrefix("notify-test"),
		clients: make([]*wsClients, 0),
	}

	t.Run("add new channel", func(t *testing.T) {
		ch := c.Add("vault", "123")
		require.NotNil(t, ch)
	})

	t.Run("find not existing channel", func(t *testing.T) {
		_, ok := c.Find("123")
		require.False(t, ok)
	})

	t.Run("find existing channel", func(t *testing.T) {
		chs, ok := c.Find("vault")
		require.True(t, ok)
		require.NotNil(t, chs)
		require.NotEmpty(t, chs)
	})

	t.Run("get channels", func(t *testing.T) {
		wsc := c.Cleanup()
		require.NotNil(t, wsc)
	})

	t.Run("notify", func(t *testing.T) {
		var got bool
		message := models.Message{
			Text:    "test message",
			Content: "this is test",
		}
		c.Add("test", "123")
		ch_out := c.Add("test", "321")
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func(in chan models.Message) {
			msg := <-in
			require.Equal(t, message, msg)
			got = true
			wg.Done()
		}(ch_out)
		c.Notify("test", "123", message)
		wg.Wait()
		require.True(t, got)
	})

	t.Run("check mutex", func(t *testing.T) {
		mux := GetMutex()
		require.NotNil(t, mux)
	})

	t.Run("cross package channels", func(t *testing.T) {
		chs := GetMsgChan()
		require.NotNil(t, chs)
	})
}
