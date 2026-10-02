/**
 * tab_perkey.js - Individual Per-Key RGB Canvas Painter & YAML Manager.
 */

class TabPerKey {
    constructor(app) {
        this.app = app;
        this.activeColor = '#00FFCC';
        this.keyColorMap = new Map(); // keyId -> hexColor
        this.isMouseDown = false;
    }

    init() {
        this.colorPicker = document.getElementById('perkey-color-picker');
        this.colorHex = document.getElementById('perkey-color-hex');
        this.applyBtn = document.getElementById('btn-apply-perkey');
        this.clearBtn = document.getElementById('btn-perkey-clear');
        this.exportBtn = document.getElementById('btn-perkey-export');
        this.importBtn = document.getElementById('btn-perkey-import');

        // Color input changes
        this.colorPicker.addEventListener('input', (e) => {
            this.activeColor = e.target.value;
            this.colorHex.value = this.activeColor.toUpperCase();
        });
        this.colorHex.addEventListener('change', (e) => {
            this.activeColor = e.target.value;
            this.colorPicker.value = this.activeColor;
        });

        // Setup mouse drag painting on canvas keycaps
        this.setupCanvasPainting();

        // Zone Preset Buttons
        document.getElementById('preset-wasd').addEventListener('click', () => this.applyPresetWasd());
        document.getElementById('preset-numpad').addEventListener('click', () => this.applyPresetNumpad());
        document.getElementById('preset-rainbow').addEventListener('click', () => this.applyPresetRainbow());

        // Actions
        this.applyBtn.addEventListener('click', () => this.applyPerKeyRgb());
        this.clearBtn.addEventListener('click', () => this.clearAll());
        this.exportBtn.addEventListener('click', () => this.exportYaml());
        this.importBtn.addEventListener('click', () => this.importYaml());
    }

    setupCanvasPainting() {
        const stage = document.getElementById('keyboard-stage-container');
        
        stage.addEventListener('mousedown', () => { this.isMouseDown = true; });
        window.addEventListener('mouseup', () => { this.isMouseDown = false; });

        stage.addEventListener('mouseover', (e) => {
            if (this.app.activeTab !== 'perkey' || !this.isMouseDown) return;
            const btn = e.target.closest('.keycap');
            if (btn && btn.dataset.id) {
                this.paintKey(btn.dataset.id, this.activeColor);
            }
        });
    }

    paintKey(keyId, color) {
        this.keyColorMap.set(keyId, color);
        this.app.canvas.setKeyLed(keyId, color);
        this.updateCommand();
    }

    clearAll() {
        this.keyColorMap.clear();
        this.app.canvas.clearAllLeds();
        this.app.composer.setCommand('f108-pro off');
    }

    applyPresetWasd() {
        this.clearAll();
        // Paint WASD red
        ['W', 'A', 'S', 'D'].forEach(k => this.paintKey(k, '#FF0033'));
        // Paint Arrows cyan
        ['UP', 'DOWN', 'LEFT', 'RIGHT'].forEach(k => this.paintKey(k, '#00F0FF'));
        // Paint 1-4 yellow
        ['1', '2', '3', '4'].forEach(k => this.paintKey(k, '#FFD700'));
    }

    applyPresetNumpad() {
        this.clearAll();
        const numKeys = ['NUMLOCK', 'KP_DIVIDE', 'KP_MULTIPLY', 'KP_MINUS', 'KP_PLUS', 'KP_ENTER',
                         'KP_0', 'KP_1', 'KP_2', 'KP_3', 'KP_4', 'KP_5', 'KP_6', 'KP_7', 'KP_8', 'KP_9', 'KP_DOT'];
        numKeys.forEach(k => this.paintKey(k, '#10B981'));
    }

    applyPresetRainbow() {
        this.clearAll();
        const colors = ['#FF0000', '#FF7F00', '#FFFF00', '#00FF00', '#0000FF', '#4B0082', '#9400D3'];
        if (this.app.canvas.keysData) {
            this.app.canvas.keysData.keys.forEach((k, idx) => {
                const c = colors[idx % colors.length];
                this.paintKey(k.id, c);
            });
        }
    }

    updateCommand() {
        if (this.keyColorMap.size === 0) {
            this.app.composer.setCommand('f108-pro perkey --help');
            return;
        }

        // Compose quartet command if few keys, or YAML if many
        if (this.keyColorMap.size <= 6) {
            const parts = [];
            this.keyColorMap.forEach((hex, id) => {
                const rgb = this.app.hexToRgb(hex);
                parts.push(`${id} ${rgb.r} ${rgb.g} ${rgb.b}`);
            });
            this.app.composer.setCommand(`f108-pro perkey ${parts.join(' ')}`);
        } else {
            this.app.composer.setCommand(`f108-pro perkey custom_profile.yaml`);
        }
    }

    generateYaml() {
        let yaml = "# AULA F108 Pro Per-Key RGB Configuration\nbrightness: 5\nall: [0, 0, 0]\nkeys:\n";
        this.keyColorMap.forEach((hex, id) => {
            const rgb = this.app.hexToRgb(hex);
            yaml += `  ${id}: [${rgb.r}, ${rgb.g}, ${rgb.b}]\n`;
        });
        return yaml;
    }

    async applyPerKeyRgb() {
        if (this.keyColorMap.size === 0) return;
        try {
            const yaml = this.generateYaml();
            const tmpPath = "/tmp/f108_perkey_active.yaml";
            await window.hostBridge.invoke('write_file', { path: tmpPath, content: yaml });
            this.app.composer.setCommand(`f108-pro perkey ${tmpPath}`);
            await this.app.composer.runCurrentCommand();
        } catch (e) {
            console.error('Failed to apply per-key RGB:', e);
        }
    }

    async exportYaml() {
        try {
            const fileRes = await window.hostBridge.invoke('save_file_dialog', {
                title: 'Export Per-Key RGB Profile',
                default_name: 'f108_rgb_profile.yaml'
            });
            if (fileRes && fileRes.path) {
                const yaml = this.generateYaml();
                await window.hostBridge.invoke('write_file', { path: fileRes.path, content: yaml });
                this.app.composer.appendLog(`[OK] Per-Key RGB profile exported to: ${fileRes.path}\n`, 'info');
            }
        } catch (e) {
            console.error('Export failed:', e);
        }
    }

    async importYaml() {
        try {
            const fileRes = await window.hostBridge.invoke('open_file_dialog', {
                title: 'Import Per-Key RGB Profile',
                filter: 'yaml'
            });
            if (fileRes && fileRes.path) {
                this.app.composer.setCommand(`f108-pro perkey ${fileRes.path}`);
                await this.app.composer.runCurrentCommand();
            }
        } catch (e) {
            console.error('Import failed:', e);
        }
    }
}

window.TabPerKey = TabPerKey;
