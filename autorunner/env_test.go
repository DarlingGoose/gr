package autorunner

import "testing"

func TestRecommendedWineEnv(t *testing.T) {
	env := RecommendedWineEnv(WineEnvConfig{
		Lang:                 "en_US.UTF-8",
		WineDebug:            "-all",
		DisableWineMenuBuild: true,
		QuietDXVKLogs:        true,
		UnattendedWinetricks: true,
		Extra: []string{
			"LANG=ja_JP.UTF-8",
			"PROTON_LOG=0",
			"",
		},
	})

	want := []string{
		"LANG=ja_JP.UTF-8",
		"WINEDEBUG=-all",
		"WINEDLLOVERRIDES=winemenubuilder.exe=d",
		"DXVK_LOG_LEVEL=none",
		"VKD3D_DEBUG=none",
		"WINETRICKS_OPT_UNATTENDED=1",
		"PROTON_LOG=0",
	}

	if len(env) != len(want) {
		t.Fatalf("len(env) = %d, want %d: %#v", len(env), len(want), env)
	}

	for i := range want {
		if env[i] != want[i] {
			t.Fatalf("env[%d] = %q, want %q", i, env[i], want[i])
		}
	}
}
