package tls

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/keeper/pkg/helpers"
)

func TestTLSinit(t *testing.T) {
	Prepare()
	require.True(t, helpers.FileExists("./ssl"))
	require.True(t, helpers.FileExists("./ssl/server.pem"))
	require.True(t, helpers.FileExists("./ssl/server.crt"))
	err := os.Remove("./ssl/server.pem")
	require.NoError(t, err)
	err = os.Remove("./ssl/server.crt")
	require.NoError(t, err)
	err = os.RemoveAll("./ssl")
	require.NoError(t, err)
}
