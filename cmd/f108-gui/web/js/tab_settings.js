/**
 * tab_settings.js - Preferences, Themes, Language & Driver Configuration.
 */

class TabSettings {
    constructor(app) {
        this.app = app;
    }

    async init() {
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

        this.sleep = document.getElementById('settings-sleep');
        this.response = document.getElementById('settings-response');
        this.note = document.getElementById('settings-state-note');
        this.applyButton = document.getElementById('settings-apply');
        const t = (key, fallback) => window.i18n.t(`settings.${key}`, fallback);
        try {
            const state = await window.hostBridge.invoke('settings_state');
            this.sleep.value = String(state.sleep);
            this.response.value = String(state.response);
            this.note.textContent = state.known
                ? t('saved_local', 'Último valor enviado desde esta GUI.')
                : t('unknown_state', 'Sin ajuste guardado; se muestran valores iniciales sugeridos.');
        } catch (e) {
            this.note.textContent = e.message;
        }

        this.applyButton.addEventListener('click', async () => {
            this.applyButton.disabled = true;
            try {
                const result = await window.hostBridge.invoke('settings_apply', {
                    sleep: Number(this.sleep.value), response: Number(this.response.value),
                });
                this.note.textContent = result.confirmed
                    ? t('applied', 'Ajustes confirmados por el teclado.')
                    : t('no_ack', 'Paquete enviado sin confirmación; pulsa una tecla y comprueba el resultado.');
                this.app.composer.appendLog(`[AJUSTES] ${this.note.textContent}\n`, result.confirmed ? 'info' : 'warning');
            } catch (e) {
                this.note.textContent = e.message;
                this.app.composer.appendLog(`[ERROR] ${e.message}\n`, 'error');
            } finally {
                this.applyButton.disabled = false;
            }
        });
    }
}

window.TabSettings = TabSettings;
