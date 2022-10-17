package cli

import (
	"fmt"
	"strings"

	"github.com/AlekseyKas/keeper/pkg/helpers"
)

// executor - parse user input and execute functionality
func (c *CLI) executor(in string) {
	prefix := strings.Split(livePrefixState.livePrefix, "/")
	switch in {
	case "":
		if livePrefixState.livePrefix == "" {
			livePrefixState.livePrefix = ">>> "
		}
		livePrefixState.isEnable = true
	case "save":
		if livePrefixState.livePrefix == "" {
			livePrefixState.livePrefix = ">>> "
		}
		c.save()
	case "quit":
		c.save()
		if c.config.Mode != "standalone" {
			c.api.Close()
		}
		livePrefixState.livePrefix = "quit> "
		c.wg.Done()
	case "..":
		if len(prefix) > 1 {
			newprefix := []string{}
			for i := 0; i < len(prefix)-1; i++ {
				newprefix = append(newprefix, prefix[i])
			}
			livePrefixState.livePrefix = strings.Join(newprefix, "/") + "> "
		}
		if len(prefix) == 1 {
			livePrefixState.isEnable = false
			livePrefixState.livePrefix = ">>> "
		}
	default:
		line := strings.Split(in, " ")
		if cmd, ok := helpers.FindCommand(in); ok {
			switch cmd {
			case "timemachine":
				if livePrefixState.livePrefix == "history> " {
					c.timemachine()
					return
				} else {
					fmt.Println("no such command")
				}
			case "rollback":
				if len(line) == 1 {
					fmt.Println("invalid command format")
					return
				}
				if livePrefixState.livePrefix == "history> " {
					c.rollback(line[1])
					return
				} else {
					fmt.Println("no such command")
				}
			case "roster":
				if livePrefixState.livePrefix == "user> " {
					c.roster()
				} else {
					fmt.Println("no such command")
				}
			case "revoke":
				if len(line) == 1 {
					fmt.Println("invalid command format")
					return
				}
				if livePrefixState.livePrefix == "user> " {
					if helpers.GenHash([]byte(c.crypto.GetPublicKey())) == line[1] {
						c.log().Warn(nil, "can revoke self")
						return
					}
					c.revoke(line[1])
				} else {
					fmt.Println("no such command")
				}
			case "confirm":
				if len(line) == 1 {
					fmt.Println("invalid command format")
					return
				}
				if livePrefixState.livePrefix == "user> " {
					c.confirm(line[1])
				} else {
					fmt.Println("no such command")
				}
			case "list":
				if livePrefixState.livePrefix == "cmd> " {
					c.list()
				} else {
					fmt.Println("no such command")
				}
			case "get":
				if len(line) == 1 {
					fmt.Println("invalid command format")
					return
				}
				if livePrefixState.livePrefix == "cmd> " {
					opts := line[2:]
					c.get(line[1], opts...)
				} else {
					fmt.Println("no such command")
				}
			case "insert":
				if len(line) == 1 {
					fmt.Println("invalid command format")
					return
				}
				if livePrefixState.livePrefix == "cmd> " {
					c.insert(in)

				} else {
					fmt.Println("no such command")
				}
			case "delete":
				if livePrefixState.livePrefix == "cmd> " {
					if len(line) == 2 {
						c.delete(line[1])
					}
				} else {
					fmt.Println("no such command")
				}
			case "view":
				if livePrefixState.livePrefix == "config> " {
					c.view()
				} else {
					fmt.Println("no such command")
				}
			case "status":
				if livePrefixState.livePrefix == ">>> " || livePrefixState.livePrefix == "" {
					c.status()
					return
				} else {
					fmt.Println("no such command")
				}
			}
			return
		}
		var found bool
		newprefix := []string{}
		if len(prefix) == 1 {
			// p := strings.Split(prefix[0], ">")
			// if len(p) == 4 {
			// 	p = []string{}
			// }
			// if len(p) > 1 {
			// 	for i := 0; i <= len(prefix)-1; i++ {
			// 		newprefix = append(newprefix, prefix[i])
			// 	}
			// 	for _, t := range suggests[livePrefixState.livePrefix] {
			// 		if t.Text == in {
			// 			found = true
			// 		}
			// 	}
			// 	if found {
			// 		livePrefixState.livePrefix = strings.TrimSuffix(strings.TrimSpace(strings.Join(newprefix, "/")), ">") + "/" + strings.TrimSpace(in) + "> "
			// 	} else {
			// 		fmt.Println("no such command")
			// 	}
			// } else {
			for i := 0; i <= len(prefix)-1; i++ {
				newprefix = append(newprefix, prefix[i])
			}
			for _, t := range suggests[livePrefixState.livePrefix] {
				if t.Text == in {
					found = true
				}
			}
			if found {
				livePrefixState.livePrefix = strings.TrimSpace(in) + "> "
			} else {
				fmt.Println("no such command")
			}
		}
		// }
		if len(prefix) > 1 {
			for i := 0; i <= len(prefix)-1; i++ {
				newprefix = append(newprefix, prefix[i])
			}
			for _, t := range suggests[livePrefixState.livePrefix] {
				if t.Text == in {
					found = true
				}
			}
			if found {
				livePrefixState.livePrefix = strings.TrimSuffix(strings.TrimSpace(strings.Join(newprefix, "/")), ">") + "/" + strings.TrimSpace(in) + "> "
			} else {
				fmt.Println("no such command")
			}
		}
	}
	livePrefixState.isEnable = true
}
