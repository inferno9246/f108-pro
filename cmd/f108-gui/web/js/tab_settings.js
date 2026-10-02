/**
 * tab_settings.js - Preferences, Themes, Language & Driver Configuration.
 */

class TabSettings {
    constructor(app) {
        this.app = app;
    }

    init() {
        this.cliPathInput = document.getElementById('input-cli-path');
        
        // Load initial CLI path from backend
        window.hostBridge.invoke('get_app_info').then(info => {
            if (info && info.cli_path) {
                this.cliPathInput.value = info.cli_path;
            }
        });

        this.cliPathInput.addEventListener('change', (e) => {
            const newPath = e.target.value.trim();
            if (newPath) {
                this.app.composer.appendLog(`[SETTINGS] Configured CLI binary: ${newPath}\n`, 'info');
            }
        });
    }
}

window.TabSettings = TabSettings;
