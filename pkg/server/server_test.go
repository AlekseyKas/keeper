package server

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	apigrpc "github.com/AlekseyKas/keeper/pkg/server/api-grpc"
	mockEndpoint "github.com/AlekseyKas/keeper/pkg/server/api-grpc/mock_apigrpc"
	mockStorage "github.com/AlekseyKas/keeper/pkg/server/storage/mock_storage"
)

func TestNew(t *testing.T) {
	ks := New()
	require.NotNil(t, ks)
}

func TestStartToEnd(t *testing.T) {

	sig := make(chan struct{})
	wg := new(sync.WaitGroup)
	ctl := gomock.NewController(t)
	db := mockStorage.NewMockStorage(ctl)
	db.EXPECT().Close().AnyTimes().Return(nil)
	endpoint := mockEndpoint.NewMockAPIandGRPC(ctl)
	endpoint.EXPECT().Stop().AnyTimes().DoAndReturn(func() error {
		close(sig)
		return nil
	})
	endpoint.EXPECT().Start(gomock.Any()).DoAndReturn(func(wg *sync.WaitGroup) error {
		<-sig
		wg.Done()
		return nil
	})
	defer ctl.Finish()
	ks := &KeeperServer{
		logger: zerolog.New().WithPrefix("server-test"),
		sig:    nil,
		Wg:     wg,
		db:     db,
		ag:     []apigrpc.APIandGRPC{endpoint},
		mux:    new(sync.Mutex),
	}
	go func() {
		errStart := ks.Start()
		require.NoError(t, errStart)
	}()
	time.Sleep(3 * time.Second)
	ks.Stop()
	require.NotNil(t, ks.log())
}
