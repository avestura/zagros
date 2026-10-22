package builtin

import "testing"

func TestParseStartAppsKeepsPlainAppIDs(t *testing.T) {
	b := []byte(`[{"Name":"Docker Desktop","AppID":"Docker.DockerForWindows.Settings"},{"Name":"Calc","AppID":"Microsoft.WindowsCalculator_8wekyb3d8bbwe!App"},{"Name":"","AppID":"x"}]`)
	got := parseStartApps(b)
	if len(got) != 2 || got[0].P != `shell:AppsFolder\Docker.DockerForWindows.Settings` {
		t.Fatalf("got %+v", got)
	}
}
