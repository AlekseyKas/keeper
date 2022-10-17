package cli

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/golang/mock/gomock"
// 	"github.com/labstack/gommon/log"
// 	"github.com/stretchr/testify/require"
// 	"github.com/AlekseyKas/keeper/pkg/client/config"
// 	mockOpenPGP "github.com/AlekseyKas/keeper/pkg/client/openpgp/mock_openpgp"
// 	mockActions "github.com/AlekseyKas/keeper/pkg/client/remote/mock_actions"
// 	mockStorage "github.com/AlekseyKas/keeper/pkg/client/storage/mock_storage"
// 	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
// )

// func clean(t *testing.T) {
// 	content, err := os.ReadDir("./openpgp")
// 	require.NoError(t, err)
// 	for _, file := range content {
// 		path := fmt.Sprintf("./openpgp/%s", file.Name())
// 		err := os.Remove(path)
// 		require.NoError(t, err)
// 		log.Debugf("%s deleted", nil, path)
// 	}
// 	err = os.RemoveAll("./openpgp")
// 	require.NoError(t, err)
// }

// func setEnv(t *testing.T) {
// 	err := os.Setenv("KEEPER_REMOTE_USERNAME", "user")
// 	require.NoError(t, err)
// 	err = os.Setenv("KEEPER_REMOTE_PASSWORD", "password")
// 	require.NoError(t, err)
// 	err = os.Setenv("KEEPER_REMOTE_URL", "https://127.0.0.1:8443")
// 	require.NoError(t, err)
// }

// func unsetEnv(t *testing.T) {
// 	err := os.Unsetenv("KEEPER_REMOTE_USERNAME")
// 	require.NoError(t, err)
// 	err = os.Unsetenv("KEEPER_REMOTE_PASSWORD")
// 	require.NoError(t, err)
// 	err = os.Unsetenv("KEEPER_REMOTE_URL")
// 	require.NoError(t, err)
// }

// func TestNew(t *testing.T) {
// 	defer clean(t)
// 	wg := new(sync.WaitGroup)
// 	got := New(wg)
// 	require.NotNil(t, got)
// }

// func TestStart(t *testing.T) {
// 	setEnv(t)
// 	defer unsetEnv(t)
// 	ctl := gomock.NewController(t)
// 	defer ctl.Finish()
// 	db := mockStorage.NewMockStorage(ctl)
// 	// actions := mockActions.NewMockActions(ctl)
// 	crypto := mockOpenPGP.NewMockOPENPGP(ctl)
// 	livePrefixState.isEnable = true
// 	livePrefixState.livePrefix = ""
// 	cli := &CLI{
// 		wg:      new(sync.WaitGroup),
// 		storage: db,
// 		logger:  zerolog.New().WithPrefix("completer-test"),
// 		// api:     actions,
// 		config: config.New(),
// 		crypto: crypto,
// 	}
// 	cli.wg.Add(1)
// 	go cli.Start()
// 	time.Sleep(10 * time.Second)
// 	fmt.Fprintln(os.Stdin, "quit")
// 	cli.wg.Wait()
// }

// func TestRemote(t *testing.T) {
// 	setEnv(t)
// 	defer unsetEnv(t)
// 	ctl := gomock.NewController(t)
// 	defer ctl.Finish()
// 	db := mockStorage.NewMockStorage(ctl)
// 	actions := mockActions.NewMockActions(ctl)
// 	crypto := mockOpenPGP.NewMockOPENPGP(ctl)
// 	livePrefixState.isEnable = true
// 	livePrefixState.livePrefix = ""
// 	cli := &CLI{
// 		wg:      new(sync.WaitGroup),
// 		storage: db,
// 		logger:  zerolog.New().WithPrefix("completer-test"),
// 		api:     actions,
// 		config:  config.New(),
// 		crypto:  crypto,
// 	}

// 	gomock.InOrder(
// 		crypto.EXPECT().GetPublicKey().Return("12345"),
// 		actions.EXPECT().Login("user", "password", "12345").Return(errors.New("internal server error")),

// 		// crypto.EXPECT().GetPublicKey().Return("12345"),
// 		// actions.EXPECT().Login("user", "password", "12345").Return(errors.New("bad request")),

// 		// crypto.EXPECT().GetPublicKey().Return("12345"),
// 		// actions.EXPECT().Login("user", "password", "12345").Return(nil),
// 		// actions.EXPECT().Register("user", "password", "12345").Return(errors.New("dummy error")),
// 	)

// 	t.Run("run 1", func(t *testing.T) { cli.remote() })
// 	// t.Run("run 2", func(t *testing.T) { cli.remote() })
// 	// t.Run("run 3", func(t *testing.T) { cli.remote() })
// }
