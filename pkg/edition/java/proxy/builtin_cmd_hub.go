package proxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.minekube.com/brigodier"
	. "go.minekube.com/common/minecraft/color"
	. "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
)

const hubCmdPermission = "gate.command.hub"
const hubName = "hub"

// command to list and connect to registered servers
func newHubCmd(proxy *Proxy) brigodier.LiteralNodeBuilder {
	return brigodier.Literal("hub").
		Requires(hasCmdPerm(proxy, hubCmdPermission)).
		// List registered server.
		Executes(command.Command(func(c *command.Context) error {
			player, ok := c.Source.(Player)
			if !ok {
				return c.Source.SendMessage(&Text{S: Style{Color: Red},
					Content: "Only players can connect to hub!"})
			}
			return connectPlayersToHub(c, proxy, hubName, player)
		}))
}

func connectPlayersToHub(c *command.Context, proxy *Proxy, serverName string, players ...Player) error {
	server := proxy.Server(serverName)
	if server == nil {
		return c.Source.SendMessage(&Text{S: Style{Color: Red},
			Content: fmt.Sprintf("Server %q doesn't exist.", serverName)})
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(),
			time.Millisecond*time.Duration(proxy.cfg.ConnectionTimeout))
		defer cancel()

		wg := new(sync.WaitGroup)
		wg.Add(len(players))
		for _, player := range players {
			go func(player Player) {
				defer wg.Done()
				player.CreateConnectionRequest(server).ConnectWithIndication(ctx)
			}(player)
		}
		wg.Wait()
	}()

	return nil
}
