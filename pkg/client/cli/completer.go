package cli

import (
	"strings"
	"time"

	"github.com/AlekseyKas/keeper/pkg/helpers"
	prompt "github.com/c-bata/go-prompt"
)

// completer - live TUI input suggestion
func (c *CLI) completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	line := in.Text
	switch livePrefixState.livePrefix {
	case "user> ":
		lst, err := c.api.ListPGP()
		if err != nil {
			c.log().Error(err, "pgp public key list can not be retrieved")
			return nil
		}
		checksums := []string{}
		for _, key := range lst {
			checksums = append(checksums, helpers.GenHash([]byte(key.Publickey)))
		}
		if strings.Contains(line, "revoke") {
			if len(checksums) != 0 {
				ss := []prompt.Suggest{}
				for _, v := range checksums {
					ss = append(ss, prompt.Suggest{Text: v})
				}
				return prompt.FilterHasPrefix(ss, w, true)
			}
		}
		if strings.Contains(line, "confirm") {
			if len(checksums) != 0 {
				ss := []prompt.Suggest{}
				for _, v := range checksums {
					ss = append(ss, prompt.Suggest{Text: v})
				}
				return prompt.FilterHasPrefix(ss, w, true)
			}
		}
	case "history> ":
		if strings.Contains(line, "rollback") {
			list, err := c.api.Versions()
			if err != nil {
				s = suggests[livePrefixState.livePrefix]
				return prompt.FilterHasPrefix(s, w, true)
			}
			if len(list) != 0 {
				ss := []prompt.Suggest{}
				for _, v := range list {
					ss = append(ss, prompt.Suggest{Text: v.Hash, Description: v.Date.Format(time.RFC822)})
				}
				return prompt.FilterHasPrefix(ss, w, true)
			}
		}
	case "cmd> ":
		if strings.Contains(line, "get") {
			list := c.storage.ListSecrets()
			if len(list) != 0 {
				ss := []prompt.Suggest{}
				for k, v := range list {
					ss = append(ss, prompt.Suggest{Text: k, Description: v})
				}
				return prompt.FilterHasPrefix(ss, w, true)
			}
		}
		if strings.Contains(line, "delete") {
			list := c.storage.ListSecrets()
			if len(list) != 0 {
				ss := []prompt.Suggest{}
				for k, v := range list {
					ss = append(ss, prompt.Suggest{Text: k, Description: v})
				}
				return prompt.FilterHasPrefix(ss, w, true)
			}
		}
		if strings.Contains(line, "insert") {
			args := strings.Split(line, " ")
			if len(args) == 3 {
				return prompt.FilterHasPrefix(insertSuggest, w, true)
			}
			if len(args) > 3 {
				switch args[2] {
				case "otp":
					return prompt.FilterHasPrefix(modifySuggest["otp"], w, true)
				case "userpass":
					return prompt.FilterHasPrefix(modifySuggest["userpass"], w, true)
				case "creditcard":
					return prompt.FilterHasPrefix(modifySuggest["creditcard"], w, true)
				case "anytext":
					return prompt.FilterHasPrefix(modifySuggest["anytext"], w, true)
				case "anybinary":
					return prompt.FilterHasPrefix(modifySuggest["anybinary"], w, true)
				}
				return prompt.FilterHasPrefix(insertSuggest, w, true)
			}
		}
	}
	if _, ok := suggests[livePrefixState.livePrefix]; ok {
		s = suggests[livePrefixState.livePrefix]
		return prompt.FilterHasPrefix(s, w, true)
	}
	return prompt.FilterHasPrefix(s, w, true)
}

// changelivePrefix - chage current TUI prefix
func changelivePrefix() (string, bool) {
	return livePrefixState.livePrefix, livePrefixState.isEnable
}
