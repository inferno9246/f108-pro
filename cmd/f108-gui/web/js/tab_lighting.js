/**
 * tab_lighting.js - Global RGB lighting over the 2.4G dongle.
 *
 * Derived from f108-pro-gui (MIT, see THIRD_PARTY.md). Instead of building a CLI
 * string, it sends a JSON action to the Go server.
 */

class TabLighting {
    constructor(app) {
        this.app = app;
        this.directionModes = [10, 11, 12, 16, 18]; // Scrolling, Rolling, Rotating, Flowing, Tilt
        this.rainbowModes = [6, 8];                  // Colourful, Spectrum (arcoiris forzado)
    }

    init() {
        this.modeSelect = document.getElementById('lighting-mode-select');
        this.sliderBrightness = document.getElementById('slider-brightness');
        this.sliderSpeed = document.getElementById('slider-speed');
        this.colorPicker = document.getElementById('lighting-color-picker');
        this.applyBtn = document.getElementById('btn-apply-lighting');

        this.buildExtraControls();
        this.modeSelect.addEventListener('change', () => this.refreshExtras());
        this.rainbowCheck.addEventListener('change', () => this.updateCommand());
        this.dirSelect.addEventListener('change', () => this.updateCommand());
        this.applyBtn.addEventListener('click', () => this.apply());
        this.refreshExtras();
    }

    buildExtraControls() {
        const grid = this.colorPicker.closest('.controls-grid');
        const mkGroup = () => {
            const g = document.createElement('div');
            g.className = 'form-group';
            return g;
        };
        const mkSpan = (key, text) => {
            const s = document.createElement('span');
            s.setAttribute('data-i18n', key);
            s.textContent = text;
            return s;
        };

        this.rainbowGroup = mkGroup();
        const rl = document.createElement('label');
        rl.style.display = 'flex';
        rl.style.alignItems = 'center';
        rl.style.gap = '8px';
        this.rainbowCheck = document.createElement('input');
        this.rainbowCheck.type = 'checkbox';
        rl.appendChild(this.rainbowCheck);
        rl.appendChild(mkSpan('lighting.colorful', 'Multicolor (arcoiris)'));
        this.rainbowGroup.appendChild(rl);

        this.dirGroup = mkGroup();
        const dl = document.createElement('label');
        dl.appendChild(mkSpan('lighting.direction', 'Direccion'));
        this.dirSelect = document.createElement('select');
        [['0', 'lighting.dir_normal', 'Normal'], ['1', 'lighting.dir_reverse', 'Invertida']].forEach(([v, k, t]) => {
            const o = document.createElement('option');
            o.value = v;
            o.setAttribute('data-i18n', k);
            o.textContent = t;
            this.dirSelect.appendChild(o);
        });
        this.dirGroup.appendChild(dl);
        this.dirGroup.appendChild(this.dirSelect);

        grid.appendChild(this.rainbowGroup);
        grid.appendChild(this.dirGroup);
        if (window.i18n) window.i18n.applyTranslations(grid);
    }

    modeIndex() {
        return this.app.lightingModes.indexOf(this.modeSelect.value);
    }

    refreshExtras() {
        const m = this.modeIndex();
        const forced = this.rainbowModes.includes(m);
        if (forced) this.rainbowCheck.checked = true;
        this.rainbowCheck.disabled = forced;
        this.dirGroup.style.display = this.directionModes.includes(m) ? '' : 'none';
        this.updateCommand();
    }

    payload() {
        const rgb = this.app.hexToRgb(this.colorPicker.value);
        const m = this.modeIndex();
        return {
            mode: Math.max(m, 0),
            r: rgb.r, g: rgb.g, b: rgb.b,
            brightness: parseInt(this.sliderBrightness.value, 10),
            speed: parseInt(this.sliderSpeed.value, 10),
            direction: this.directionModes.includes(m) ? parseInt(this.dirSelect.value, 10) : 0,
            colorful: this.rainbowModes.includes(m) || this.rainbowCheck.checked,
        };
    }

    // Comando equivalente del CLI por cable (solo informativo; no se ejecuta).
    updateCommand() {
        if (!this.app.composer) return;
        const p = this.payload();
        const name = this.modeSelect.value || 'Static';
        const color = p.colorful ? 'colorful' : `${p.r} ${p.g} ${p.b}`;
        const dir = p.direction ? ` -d ${p.direction}` : '';
        this.app.composer.setCommand(`f108-pro light ${name} ${p.brightness} ${p.speed} ${color}${dir}`);
    }

    log(msg, type) {
        if (this.app.composer && this.app.composer.appendLog) {
            this.app.composer.appendLog(msg + '\n', type || 'info');
        } else {
            console.log(msg);
        }
    }

    async apply() {
        const p = this.payload();
        this.applyBtn.disabled = true;
        try {
            const res = await window.hostBridge.invoke('light_apply', p);
            if (res && res.confirmed === false) {
                this.log(`[AVISO] Iluminacion enviada (${this.modeSelect.value}), pero el teclado no la confirmo.`);
            } else {
                this.log(`[OK] Iluminacion aplicada: ${this.modeSelect.value}`);
            }
        } catch (e) {
            this.log(`[ERROR] ${e.message}`, 'error');
        } finally {
            this.applyBtn.disabled = false;
        }
    }
}

window.TabLighting = TabLighting;
