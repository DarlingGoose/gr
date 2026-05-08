package autorunner

import "strings"

type WineEnvConfig struct {
	Lang                 string
	LCAll                string
	WineDebug            string
	DisableWineMenuBuild bool
	QuietDXVKLogs        bool
	UnattendedWinetricks bool
	Extra                []string
}

func DefaultWineEnvConfig() WineEnvConfig {
	return WineEnvConfig{
		Lang:                 "C.UTF-8",
		LCAll:                "C.UTF-8",
		WineDebug:            "-all",
		DisableWineMenuBuild: true,
		QuietDXVKLogs:        true,
		UnattendedWinetricks: true,
	}
}

func RecommendedWineEnv(cfg WineEnvConfig) []string {
	env := make([]string, 0, 8+len(cfg.Extra))

	if cfg.Lang != "" {
		env = appendEnv(env, "LANG", cfg.Lang)
	}

	if cfg.LCAll != "" {
		env = appendEnv(env, "LC_ALL", cfg.LCAll)
	}

	if cfg.WineDebug != "" {
		env = appendEnv(env, "WINEDEBUG", cfg.WineDebug)
	}

	if cfg.DisableWineMenuBuild {
		env = appendEnv(env, "WINEDLLOVERRIDES", "winemenubuilder.exe=d")
	}

	if cfg.QuietDXVKLogs {
		env = appendEnv(env, "DXVK_LOG_LEVEL", "none")
		env = appendEnv(env, "VKD3D_DEBUG", "none")
	}

	if cfg.UnattendedWinetricks {
		env = appendEnv(env, "WINETRICKS_OPT_UNATTENDED", "1")
	}

	for _, extra := range cfg.Extra {
		if strings.TrimSpace(extra) == "" {
			continue
		}
		env = upsertEnvSpec(env, extra)
	}

	return env
}

func appendEnv(env []string, key, value string) []string {
	return upsertEnvSpec(env, key+"="+value)
}

func upsertEnvSpec(env []string, spec string) []string {
	key, _, ok := strings.Cut(spec, "=")
	if !ok || key == "" {
		return append(env, spec)
	}

	prefix := key + "="
	for i, existing := range env {
		if strings.HasPrefix(existing, prefix) {
			env[i] = spec
			return env
		}
	}

	return append(env, spec)
}
