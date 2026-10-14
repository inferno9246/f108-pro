/**
 * tab_perkey.js - Luz RGB tecla por tecla por el dongle 2.4G.
 *
 * Derived from f108-pro-gui (MIT, see THIRD_PARTY.md). En lugar de armar comandos de texto,
 * envia al servidor un JSON {brightness, keys: {indice_de_luz: "#rrggbb"}}.
 */

class TabPerKey {
    constructor(app) {
        this.app = app;
        this.activeColor = '#00FFCC';
        this.colors = new Map();   // indice de luz -> "#rrggbb"
        this.brightness = 5;
        this.isMouseDown = false;
        this.names = new Map();    // nombre de tecla (del servidor) -> indice de luz
    }

    t(key, fallback) {
        return (window.i18n && window.i18n.t(key, fallback)) || fallback;
    }

    log(msg, type) {
        if (this.app.composer && this.app.composer.appendLog) {
            this.app.composer.appendLog(msg + '\n', type || 'info');
        } else {
            console.log(msg);
        }
    }

    keys() {
        return (this.app.canvas && this.app.canvas.keysData) ? this.app.canvas.keysData.keys : [];
    }

    keyById(id) { return this.keys().find(k => k.id === id) || null; }
    keyByIndex(idx) { return this.keys().find(k => k.light_index === idx) || null; }

    async init() {
        this.picker = document.getElementById('perkey-color-picker');
        this.hex = document.getElementById('perkey-color-hex');
        this.applyBtn = document.getElementById('btn-apply-perkey');
        this.clearBtn = document.getElementById('btn-perkey-clear');

        // Exportar/importar YAML todavia no esta disponible en esta version.
        ['btn-perkey-export', 'btn-perkey-import'].forEach(id => {
            const b = document.getElementById(id);
            if (b) b.style.display = 'none';
        });

        this.buildExtras();

        this.picker.addEventListener('input', (e) => {
            this.activeColor = e.target.value.toUpperCase();
            this.hex.value = this.activeColor;
        });
        this.hex.addEventListener('change', (e) => {
            if (/^#[0-9a-fA-F]{6}$/.test(e.target.value)) {
                this.activeColor = e.target.value.toUpperCase();
                this.picker.value = this.activeColor;
            }
        });
        document.getElementById('preset-wasd').addEventListener('click', () => this.presetWasd());
        document.getElementById('preset-numpad').addEventListener('click', () => this.presetNumpad());
        document.getElementById('preset-rainbow').addEventListener('click', () => this.presetRainbow());
        this.clearBtn.addEventListener('click', () => this.clearAll());
        this.fillBtn.addEventListener('click', () => this.fillAll());
        this.applyBtn.addEventListener('click', () => this.apply());
        this.brightInput.addEventListener('input', () => {
            this.brightness = parseInt(this.brightInput.value, 10);
            this.brightVal.textContent = String(this.brightness);
        });

        this.setupCanvasPainting();
        try {
            await this.loadNames();
            await this.loadState();
        } catch (e) {
            this.log(`[ERROR] ${e.message}`, 'error');
        }
    }

    buildExtras() {
        const grid = this.picker.closest('.controls-grid');
        const g = document.createElement('div');
        g.className = 'form-group';
        g.innerHTML = `
            <label><span data-i18n="perkey.brightness">Brillo (0 - 5)</span>: <span class="slider-val" id="perkey-bright-val">5</span></label>
            <div class="slider-group"><input type="range" id="perkey-bright" min="0" max="5" value="5"></div>`;
        grid.appendChild(g);
        this.brightInput = g.querySelector('#perkey-bright');
        this.brightVal = g.querySelector('#perkey-bright-val');

        this.fillBtn = document.createElement('button');
        this.fillBtn.className = 'btn btn-secondary';
        this.fillBtn.id = 'btn-perkey-fill';
        this.fillBtn.setAttribute('data-i18n', 'perkey.fill_all');
        this.fillBtn.textContent = this.t('perkey.fill_all', 'Pintar todas las teclas');
        this.clearBtn.parentElement.insertBefore(this.fillBtn, this.clearBtn);
        if (window.i18n) window.i18n.applyTranslations(grid.parentElement);
    }

    async loadNames() {
        const res = await fetch('/api/keys', { headers: { 'X-Token': window.hostBridge.token } });
        if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
        (await res.json()).forEach(k => this.names.set(k.name, k.index));
    }

    async loadState() {
        const st = await window.hostBridge.invoke('perkey_state', {});
        this.brightness = Math.min(5, Math.max(0, st.brightness | 0));
        this.brightInput.value = String(this.brightness);
        this.brightVal.textContent = String(this.brightness);
        Object.entries(st.keys || {}).forEach(([idx, color]) => {
            const k = this.keyByIndex(parseInt(idx, 10));
            if (k) this.paintKey(k.id, color);
        });
    }

    setupCanvasPainting() {
        const stage = document.getElementById('keyboard-stage-container');
        if (!stage) return;
        stage.addEventListener('mousedown', () => { this.isMouseDown = true; });
        window.addEventListener('mouseup', () => { this.isMouseDown = false; });
        stage.addEventListener('mouseover', (e) => {
            if (!this.isMouseDown || this.app.activeTab !== 'perkey') return;
            const btn = e.target.closest('.keycap');
            if (btn && !btn.disabled) this.paintKey(btn.dataset.id, this.activeColor);
        });
    }

    // Se llama desde AppMain.onKeySelected (clic) y al arrastrar sobre las teclas.
    paintKey(keyId, color) {
        const k = this.keyById(keyId);
        if (!k) return;
        const hex = color.toLowerCase();
        this.colors.set(k.light_index, hex);
        this.app.canvas.setKeyLed(keyId, hex);
    }

    paintNames(names, color) {
        names.forEach(n => {
            const idx = this.names.get(n);
            const k = idx === undefined ? null : this.keyByIndex(idx);
            if (k) this.paintKey(k.id, color);
        });
    }

    clearAll() {
        this.colors.clear();
        this.app.canvas.clearAllLeds();
    }

    fillAll() {
        const valid = new Set(this.names.values());
        this.keys().forEach(k => {
            if (valid.has(k.light_index)) this.paintKey(k.id, this.activeColor);
        });
    }

    presetWasd() {
        this.clearAll();
        this.paintNames(['w', 'a', 's', 'd'], '#FF0033');
        this.paintNames(['up', 'down', 'left', 'right'], '#00F0FF');
        this.paintNames(['1', '2', '3', '4'], '#FFD700');
    }

    presetNumpad() {
        this.clearAll();
        const names = ['numlock', 'numslash', 'numstar', 'numminus', 'numplus', 'numenter', 'numdot'];
        for (let n = 0; n <= 9; n++) names.push('num' + n);
        this.paintNames(names, '#10B981');
    }

    presetRainbow() {
        this.clearAll();
        const valid = new Set(this.names.values());
        this.keys().forEach(k => {
            if (valid.has(k.light_index)) this.paintKey(k.id, this.hslToHex(Math.round(300 * k.rect.x / 800), 100, 50));
        });
    }

    hslToHex(h, s, l) {
        s /= 100; l /= 100;
        const a = s * Math.min(l, 1 - l);
        const f = (n) => {
            const k = (n + h / 30) % 12;
            const c = l - a * Math.max(-1, Math.min(k - 3, 9 - k, 1));
            return Math.round(255 * c).toString(16).padStart(2, '0');
        };
        return `#${f(0)}${f(8)}${f(4)}`;
    }

    updateCommand() {
        if (this.app.composer) this.app.composer.setCommand('f108-pro perkey --help');
    }

    async apply() {
        const keys = {};
        this.colors.forEach((hex, idx) => { keys[String(idx)] = hex; });
        this.applyBtn.disabled = true;
        try {
            await window.hostBridge.invoke('perkey_apply', { brightness: this.brightness, keys });
            this.log(`[OK] ${this.t('perkey.applied', 'Luz por tecla aplicada')} (${this.colors.size} ${this.t('perkey.keys_word', 'teclas')})`);
        } catch (e) {
            this.log(`[ERROR] ${e.message}`, 'error');
        } finally {
            this.applyBtn.disabled = false;
        }
    }
}

window.TabPerKey = TabPerKey;
