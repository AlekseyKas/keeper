package cli

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/keeper/pkg/client/config"
	mockOpenPGP "github.com/AlekseyKas/keeper/pkg/client/openpgp/mock_openpgp"
	mockActions "github.com/AlekseyKas/keeper/pkg/client/remote/mock_actions"
	mockStorage "github.com/AlekseyKas/keeper/pkg/client/storage/mock_storage"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
)

func TestCompleter(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	db := mockStorage.NewMockStorage(ctl)
	actions := mockActions.NewMockActions(ctl)
	crypto := mockOpenPGP.NewMockOPENPGP(ctl)

	cli := &CLI{
		wg:      new(sync.WaitGroup),
		storage: db,
		logger:  zerolog.New().WithPrefix("completer-test"),
		api:     actions,
		config:  config.New(),
		crypto:  crypto,
	}
	cli.config.Mode = "client-server"
	cli.wg.Add(1)

	var errDummy = errors.New("dummy error")
	testTime, err := time.Parse("2006-02-01", "2022-01-01")
	require.NoError(t, err)

	gomock.InOrder(
		actions.EXPECT().ListPGP().Return(nil, errDummy),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: testTime, Publickey: "pubkey", Confirmed: true}}, nil),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: testTime, Publickey: "pubkey", Confirmed: true}}, nil),
		actions.EXPECT().Versions().Return(nil, errDummy),
		actions.EXPECT().Versions().Return([]models.Version{{Date: testTime, Hash: "1234"}}, nil),
		db.EXPECT().ListSecrets().Return(map[string]string{"test": "test"}),
		db.EXPECT().ListSecrets().Return(map[string]string{"test": "test"}),
	)

	type test struct {
		name   string
		text   string
		prefix string
		want   []prompt.Suggest
	}
	tests := []test{
		{
			name:   "root",
			text:   "",
			prefix: ">>> ",
			want:   initSuggest[">>> "],
		},
		{
			name:   "user",
			text:   "revoke",
			prefix: "user> ",
			want:   nil,
		},
		{
			name:   "user",
			text:   "revoke",
			prefix: "user> ",
			want:   []prompt.Suggest{{Text: "b84b25628f800e36925811aa24aaf28c9f827333d2df990762b5c3a86eff7c9b", Description: ""}},
		},
		{
			name:   "user",
			text:   "confirm",
			prefix: "user> ",
			want:   []prompt.Suggest{{Text: "b84b25628f800e36925811aa24aaf28c9f827333d2df990762b5c3a86eff7c9b", Description: ""}},
		},
		{
			name:   "versions",
			text:   "rollback",
			prefix: "history> ",
			want:   nil,
		},
		{
			name:   "versions",
			text:   "rollback",
			prefix: "history> ",
			want:   []prompt.Suggest{{Text: "1234", Description: testTime.Format(time.RFC822)}},
		},
		{
			name:   "cmd",
			text:   "get ",
			prefix: "cmd> ",
			want:   []prompt.Suggest{{Text: "test", Description: "test"}},
		},
		{
			name:   "cmd",
			text:   "delete ",
			prefix: "cmd> ",
			want:   []prompt.Suggest{{Text: "test", Description: "test"}},
		},
		{
			name:   "cmd",
			text:   "insert test  ",
			prefix: "cmd> ",
			want:   insertSuggest,
		},
		{
			name:   "cmd",
			text:   "insert test anytext ",
			prefix: "cmd> ",
			want:   modifySuggest["anytext"],
		},
		{
			name:   "cmd",
			text:   "insert test anybinary ",
			prefix: "cmd> ",
			want:   modifySuggest["anybinary"],
		},
		{
			name:   "cmd",
			text:   "insert test userpass ",
			prefix: "cmd> ",
			want:   modifySuggest["userpass"],
		},
		{
			name:   "cmd",
			text:   "insert test otp ",
			prefix: "cmd> ",
			want:   modifySuggest["otp"],
		},
		{
			name:   "cmd",
			text:   "insert test creditcard ",
			prefix: "cmd> ",
			want:   modifySuggest["creditcard"],
		},
	}

	pD := prompt.NewDocument()
	livePrefixState.isEnable = true

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			livePrefixState.livePrefix = tt.prefix
			pD.Text = tt.text
			got := cli.completer(*pD)
			require.Equal(t, tt.want, got)
		})
	}
}
