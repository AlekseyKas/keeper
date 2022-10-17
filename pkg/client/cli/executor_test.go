package cli

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/keeper/pkg/client/config"
	mockOpenPGP "github.com/AlekseyKas/keeper/pkg/client/openpgp/mock_openpgp"
	mockActions "github.com/AlekseyKas/keeper/pkg/client/remote/mock_actions"
	"github.com/AlekseyKas/keeper/pkg/client/storage"
	mockStorage "github.com/AlekseyKas/keeper/pkg/client/storage/mock_storage"
	"github.com/AlekseyKas/keeper/pkg/client/storage/secrets"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
)

func TestUserInput(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	db := mockStorage.NewMockStorage(ctl)
	actions := mockActions.NewMockActions(ctl)
	crypto := mockOpenPGP.NewMockOPENPGP(ctl)

	var errDummy = errors.New("dummy")

	testCC, _ := secrets.NewCC("1234", "mr. cardholder", "01/01", 123)
	testValidTOTP, _ := secrets.NewOTP("TOTP", "localhost", "JBSWY3DPEHPK3PXP", "accountname", "1234", "1234")
	testValidHOTP, _ := secrets.NewOTP("HOTP", "issuer", "JBSWY3DPEHPK3PXP", "account", "1234", "1234")
	testInvalidTOTP := &secrets.OTP{Method: "TOTP", Issuer: "issuer", Secret: "1234", AccountName: "account"}
	testInvalidXOTP := &secrets.OTP{Method: "XOTP", Issuer: "issuer", Secret: "1234", AccountName: "account"}

	gomock.InOrder(
		// save actions
		db.EXPECT().Save().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("1234"),
		actions.EXPECT().Push("hello", "1234").Return(nil),

		db.EXPECT().Save().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("1234"),
		actions.EXPECT().Push("hello", "1234").Return(nil),

		db.EXPECT().Save().Return(nil, storage.ErrHashValid),

		db.EXPECT().Save().Return(nil, errDummy),

		db.EXPECT().Save().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return(nil, errDummy),

		db.EXPECT().Save().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("1234"),
		actions.EXPECT().Push("hello", "1234").Return(errDummy),

		//quit action
		db.EXPECT().Save().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("1234"),
		actions.EXPECT().Push("hello", "1234").Return(nil),
		actions.EXPECT().Close().Return(nil),

		// timemachine action
		actions.EXPECT().Versions().Return([]models.Version{}, nil),

		actions.EXPECT().Versions().Return(nil, errDummy),

		actions.EXPECT().Versions().Return([]models.Version{{Date: time.Now(), Hash: "123"}}, nil),

		// roster action
		actions.EXPECT().ListPGP().Return(nil, errDummy),

		actions.EXPECT().ListPGP().Return([]models.PGP{}, nil),

		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "PublicKey", Confirmed: false}}, nil),

		//revoke action
		crypto.EXPECT().GetPublicKey().Return("321"),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return(nil, errDummy),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: true}, {Date: time.Now(), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().RevokePGP("KEY").Return(errDummy),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: true}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().RevokePGP("KEY").Return(nil),
		crypto.EXPECT().ReloadPublicKeys([]string{"KEY2"}).Return(errDummy),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: true}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().RevokePGP("KEY").Return(nil),
		crypto.EXPECT().ReloadPublicKeys([]string{"KEY2"}).Return(nil),
		db.EXPECT().ReEncrypt().Return(nil, errDummy),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: true}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().RevokePGP("KEY").Return(nil),
		crypto.EXPECT().ReloadPublicKeys([]string{"KEY2"}).Return(nil),
		db.EXPECT().ReEncrypt().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return(nil, errDummy),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: true}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().RevokePGP("KEY").Return(nil),
		crypto.EXPECT().ReloadPublicKeys([]string{"KEY2"}).Return(nil),
		db.EXPECT().ReEncrypt().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("123"),
		actions.EXPECT().Push("hello", "123").Return(errDummy),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: true}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().RevokePGP("KEY").Return(nil),
		crypto.EXPECT().ReloadPublicKeys([]string{"KEY2"}).Return(nil),
		db.EXPECT().ReEncrypt().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("123"),
		actions.EXPECT().Push("hello", "123").Return(nil),

		crypto.EXPECT().GetPublicKey().Return("321"),
		actions.EXPECT().ListPGP().Return([]models.PGP{}, nil),

		//confirm action
		actions.EXPECT().ListPGP().Return(nil, errDummy),

		actions.EXPECT().ListPGP().Return([]models.PGP{}, nil),

		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: false}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().ConfirmPGP("KEY").Return(errDummy),

		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: false}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().ConfirmPGP("KEY").Return(nil),
		crypto.EXPECT().AddPublicKey([]byte("KEY")).Return(errDummy),

		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: false}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().ConfirmPGP("KEY").Return(nil),
		crypto.EXPECT().AddPublicKey([]byte("KEY")).Return(nil),
		db.EXPECT().ReEncrypt().Return(nil, errDummy),

		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: false}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().ConfirmPGP("KEY").Return(nil),
		crypto.EXPECT().AddPublicKey([]byte("KEY")).Return(nil),
		db.EXPECT().ReEncrypt().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return(nil, errDummy),

		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: false}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().ConfirmPGP("KEY").Return(nil),
		crypto.EXPECT().AddPublicKey([]byte("KEY")).Return(nil),
		db.EXPECT().ReEncrypt().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("123"),
		actions.EXPECT().Push("hello", "123").Return(errDummy),

		actions.EXPECT().ListPGP().Return([]models.PGP{{Date: time.Now(), Publickey: "KEY", Confirmed: false}, {Date: time.Now().Add(time.Hour), Publickey: "KEY2", Confirmed: true}}, nil),
		actions.EXPECT().ConfirmPGP("KEY").Return(nil),
		crypto.EXPECT().AddPublicKey([]byte("KEY")).Return(nil),
		db.EXPECT().ReEncrypt().Return([]byte("hello"), nil),
		crypto.EXPECT().EncryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().HashSum().Return("123"),
		actions.EXPECT().Push("hello", "123").Return(nil),

		// list action
		db.EXPECT().ListSecrets().Return(map[string]string{}),

		db.EXPECT().ListSecrets().Return(map[string]string{"test": "value"}),
		db.EXPECT().GetSecret("test").Return(&secrets.AnyText{Text: "test"}),

		//status action
		db.EXPECT().HashSum().Return("123"),

		// rollback action
		actions.EXPECT().Pull("124").Return(nil, errDummy),

		actions.EXPECT().Pull("124").Return([]byte("hello"), nil),
		crypto.EXPECT().DecryptWithKeys([]byte("hello")).Return(nil, errDummy),

		actions.EXPECT().Pull("124").Return([]byte("hello"), nil),
		crypto.EXPECT().DecryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().Load([]byte("hello")).Return(errDummy),

		actions.EXPECT().Pull("124").Return([]byte("hello"), nil),
		crypto.EXPECT().DecryptWithKeys([]byte("hello")).Return([]byte("hello"), nil),
		db.EXPECT().Load([]byte("hello")).Return(nil),

		//delete action
		db.EXPECT().DeleteSecret("1234").Return(db),

		//get action
		db.EXPECT().GetSecret("1234").Return(nil),

		db.EXPECT().GetSecret("1234").Return(secrets.NewText("hello")),

		db.EXPECT().GetSecret("1234").Return(testCC),

		db.EXPECT().GetSecret("1234").Return(secrets.NewUserPass("user", "pass")),

		db.EXPECT().GetSecret("1234").Return(testValidTOTP),

		db.EXPECT().GetSecret("1234").Return(testValidHOTP),

		db.EXPECT().GetSecret("1234").Return(nil),

		db.EXPECT().GetSecret("1234").Return(secrets.NewBinary([]byte("hello"))),

		db.EXPECT().GetSecret("1234").Return(secrets.NewBinary([]byte("hello"))),

		db.EXPECT().GetSecret("1234").Return(testInvalidTOTP),

		db.EXPECT().GetSecret("1234").Return(testInvalidXOTP),

		// insert action
		db.EXPECT().InsertSecret("testText", "test", secrets.NewText("test")),

		db.EXPECT().InsertSecret("testCC", "test", testCC),

		db.EXPECT().InsertSecret("testUP", "test", secrets.NewUserPass("user", "pass")),

		db.EXPECT().InsertSecret("testOTP", "test", testValidTOTP),

		db.EXPECT().InsertSecret("testBin", "test", secrets.NewBinary([]byte("hello\n"))),
	)

	cli := &CLI{
		wg:      new(sync.WaitGroup),
		storage: db,
		logger:  zerolog.New().WithPrefix("operations-test"),
		api:     actions,
		config:  config.New(),
		crypto:  crypto,
	}
	cli.config.Mode = "client-server"
	cli.wg.Add(1)

	type test struct {
		name         string
		in           string
		livePrefix   string
		testprefix   bool
		resultprefix string
	}

	tests := []test{
		{
			name: "empty action",
			in:   "",
		},
		{
			name:         "empty action",
			in:           "cmd",
			livePrefix:   ">>> ",
			testprefix:   true,
			resultprefix: "cmd> ",
		},
		{
			name:         "empty action",
			in:           "cmd",
			livePrefix:   "cmd/cmd> ",
			testprefix:   true,
			resultprefix: "cmd/cmd> ",
		},
		// save action
		{
			name:       "save action",
			in:         "save",
			livePrefix: ">>> ",
		},
		{
			name:       "save action",
			in:         "save",
			livePrefix: "",
		},
		{
			name:       "save action",
			in:         "save",
			livePrefix: "",
		},
		{
			name:       "save action",
			in:         "save",
			livePrefix: "",
		},
		{
			name:       "save action",
			in:         "save",
			livePrefix: "",
		},
		{
			name:       "save action",
			in:         "save",
			livePrefix: "",
		},
		//quit action
		{
			name:       "quit action",
			in:         "quit",
			livePrefix: "",
		},
		{
			name:       "up action",
			in:         "..",
			livePrefix: ">>> ",
		},
		{
			name:       "up action",
			in:         "..",
			livePrefix: "cmd> ",
		},
		{
			name:       "up action",
			in:         "..",
			livePrefix: "cmd/test> ",
		},
		// timemachine action
		{
			name:       "timemachine action",
			in:         "timemachine",
			livePrefix: "history> ",
		},
		{
			name:       "timemachine action",
			in:         "timemachine",
			livePrefix: "history> ",
		},
		{
			name:       "timemachine action",
			in:         "timemachine",
			livePrefix: "history> ",
		},
		{
			name:       "timemachine action",
			in:         "timemachine",
			livePrefix: "cmd> ",
		},
		// roster action
		{
			name:       "roster action",
			in:         "roster",
			livePrefix: "cmd> ",
		},
		{
			name:       "roster action",
			in:         "roster",
			livePrefix: "user> ",
		},
		{
			name:       "roster action",
			in:         "roster",
			livePrefix: "user> ",
		},
		{
			name:       "roster action",
			in:         "roster",
			livePrefix: "user> ",
		},
		{
			name:       "roster action",
			in:         "roster",
			livePrefix: "history> ",
		},
		// revoke action
		{
			name:       "revoke action",
			in:         "revoke",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 8d23cf6c86e834a7aa6eded54c26ce2bb2e74903538c61bdd5d2197997ab2f72",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "cmd> ",
		},
		{
			name:       "revoke action",
			in:         "revoke 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		// confirm action
		{
			name:       "confirm action",
			in:         "confirm",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "cmd> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		{
			name:       "confirm action",
			in:         "confirm 5ca24005b740717ba4f3f6bc48a230700e68c2a4b11ecedb96f169f4efaf1f21",
			livePrefix: "user> ",
		},
		// list action
		{
			name:       "list action",
			in:         "list",
			livePrefix: "cmd> ",
		},
		{
			name:       "list action",
			in:         "list",
			livePrefix: "cmd> ",
		},
		{
			name:       "list action",
			in:         "list",
			livePrefix: "user> ",
		},
		// view action
		{
			name:       "config action",
			in:         "view",
			livePrefix: "user> ",
		},
		{
			name:       "config action",
			in:         "view",
			livePrefix: "config> ",
		},
		// status action
		{
			name:       "status action",
			in:         "status",
			livePrefix: "cmd> ",
		},
		{
			name:       "status action",
			in:         "status",
			livePrefix: ">>> ",
		},
		// rollback action
		{
			name:       "rollback action",
			in:         "rollback 123",
			livePrefix: "cmd> ",
		},
		{
			name:       "rollback action",
			in:         "rollback",
			livePrefix: "history> ",
		},
		{
			name:       "rollback action",
			in:         "rollback 124",
			livePrefix: "history> ",
		},
		{
			name:       "rollback action",
			in:         "rollback 124",
			livePrefix: "history> ",
		},
		{
			name:       "rollback action",
			in:         "rollback 124",
			livePrefix: "history> ",
		},
		{
			name:       "rollback action",
			in:         "rollback 124",
			livePrefix: "history> ",
		},
		//delete action
		{
			name:       "delete action",
			in:         "delete 124",
			livePrefix: "history> ",
		},
		{
			name:       "delete action",
			in:         "delete",
			livePrefix: "cmd> ",
		},
		{
			name:       "delete action",
			in:         "delete 1234",
			livePrefix: "cmd> ",
		},
		// get action
		{
			name:       "get action",
			in:         "get",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get",
			livePrefix: "user> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234 secret.txt",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234",
			livePrefix: "cmd> ",
		},
		{
			name:       "get action",
			in:         "get 1234 1234",
			livePrefix: "user> ",
		},
		//insert action
		{
			name:       "insert action",
			in:         "insert",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testText anytext text test description test",
			livePrefix: "user> ",
		},
		{
			name:       "insert action",
			in:         "insert",
			livePrefix: "user> ",
		},
		{
			name:       "insert action",
			in:         "insert testText anytext text test description test",
			livePrefix: "cmd> ",
		},

		{
			name:       "insert action",
			in:         "insert testCC creditcard expire 01/01 cvv 123 number 1234 holder mr. cardholder description test",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testCC creditcard expire 01/01 cvv abc number 1234 holder mr. cardholder description test",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testUP userpass username user password pass description test",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testUP userpass username user password pass",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testUP",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testOTP otp method TOTP secret JBSWY3DPEHPK3PXP issuer localhost username accountname recoverycodes 1234 1234 description test",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testBin anybinary test.txt",
			livePrefix: "cmd> ",
		},
		{
			name:       "insert action",
			in:         "insert testBin anybinary path validbin.txt description test",
			livePrefix: "cmd> ",
		},
	}
	livePrefixState.isEnable = true
	f, err := os.OpenFile("validbin.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	require.NoError(t, err)
	fmt.Fprintln(f, "hello")
	f.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			livePrefixState.livePrefix = tt.livePrefix
			cli.executor(tt.in)
			if tt.testprefix {
				require.Equal(t, tt.resultprefix, livePrefixState.livePrefix)
			}
			os.Remove("secrets.db")
			os.Remove("secret.txt")
		})
	}
	os.Remove("validbin.txt")

}
