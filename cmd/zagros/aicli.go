package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/ai"
	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/ext"
)

const aiCmdUsage = `usage:
  zagros ai providers               list the providers in zagros.toml and whether a key is stored
  zagros ai key set <provider>      store the provider's API key in Windows Credential Manager (read from the prompt or stdin)
  zagros ai key delete <provider>   remove the stored key
  zagros ai test [provider]         send a tiny request to check the endpoint and key
`

func cmdAI(args []string) int {
	cfg, err := core.LoadConfig(core.ConfigPath())
	if err != nil {
		return fail(err)
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, aiCmdUsage)
		return 2
	}
	known := func(name string) bool {
		for _, p := range cfg.AI.Providers {
			if p.Name == name {
				return true
			}
		}
		return false
	}
	switch args[0] {
	case "providers", "list":
		if len(cfg.AI.Providers) == 0 {
			fmt.Println("no providers: add an [[ai.provider]] section to", core.ConfigPath())
			return 0
		}
		for i, p := range cfg.AI.Providers {
			key := "no key"
			switch {
			case ext.HasAIKey(p.Name):
				key = "key stored"
			case ai.Provider{BaseURL: p.BaseURL}.KeyOptional():
				key = "local, no key needed"
			}
			def := ""
			if p.Name == cfg.AI.Default || (cfg.AI.Default == "" && i == 0) {
				def = " (default)"
			}
			fmt.Printf("%-16s %-10s %-28s %s%s\n", p.Name, p.Kind, p.Model, key, def)
		}
	case "key":
		if len(args) != 3 || (args[1] != "set" && args[1] != "delete") {
			fmt.Fprint(os.Stderr, aiCmdUsage)
			return 2
		}
		name := args[2]
		if !known(name) {
			return fail(fmt.Errorf("%q is not an [[ai.provider]] in %s", name, core.ConfigPath()))
		}
		if args[1] == "delete" {
			if err := ext.DeleteAIKey(name); err != nil {
				return fail(err)
			}
			fmt.Println("removed the key of", name)
			return 0
		}
		fmt.Fprintf(os.Stderr, "API key for %s: ", name)
		key, err := readSecretLine()
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return fail(err)
		}
		if err := ext.SetAIKey(name, strings.TrimSpace(key)); err != nil {
			return fail(err)
		}
		fmt.Println("stored the key of", name, "in Credential Manager")
	case "test":
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		p, err := ext.ResolveAIProvider(cfg.AI, name)
		if err != nil {
			return fail(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		res, err := ai.Complete(ctx, http.DefaultClient, p, ai.Request{MaxTokens: 16, Messages: []ai.Message{{Role: "user", Content: "Reply with the single word: ok"}}}, 1<<20)
		if err != nil {
			return fail(err)
		}
		fmt.Printf("%s answered (%s, %d tokens): %q\n", p.Name, res.Model, res.Usage.Total(), strings.TrimSpace(res.Text))
	default:
		fmt.Fprint(os.Stderr, aiCmdUsage)
		return 2
	}
	return 0
}

func readLine() (string, error) {
	s, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && s == "" {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}
