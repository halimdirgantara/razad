package install

import "fmt"

// renderUnitFile generates the systemd unit file content for the razad
// daemon. The unit is intentionally minimal: Type=simple, Restart=on-failure,
// dynamic user, journald logging.
func renderUnitFile(opts Options) string {
	return fmt.Sprintf(`# razad-daemon.service
# Managed by the razad installer. Do not edit by hand — re-run the installer
# to regenerate this file with current defaults.

[Unit]
Description=Razad server management daemon
Documentation=https://github.com/halimdirgantara/razad
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=%s
Group=%s
ExecStart=%s
WorkingDirectory=%s
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

# Hardening (defence in depth — the daemon already validates inputs).
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=%s
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

# Journald picks up stdout/stderr automatically.
StandardOutput=journal
StandardError=journal
SyslogIdentifier=razad-daemon

[Install]
WantedBy=multi-user.target
`, opts.User, opts.Group, opts.BinaryPath, opts.DataDir, opts.DataDir)
}
