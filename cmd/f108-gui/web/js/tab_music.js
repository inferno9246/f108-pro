/**
 * tab_music.js - Music mode over the 2.4G dongle (packet 0B 1C).
 *
 * The Go server captures the system audio, computes 21 spectrum bands and streams them
 * to the keyboard. Part of the layout follows f108-pro-gui (MIT, see THIRD_PARTY.md).
 */

class TabMusic {
    constructor(app) {
        this.app = app;
        this.running = false;
        this.poll = null;
        this.pushTimer = null;
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

    init() {
        this.buildNav();
        this.buildPane();
    }

    buildNav() {
        const ref = document.querySelector('.nav-item[data-tab="lighting"]');
        if (!ref || document.querySelector('.nav-item[data-tab="music"]')) return;
        const nav = ref.cloneNode(true);
        nav.removeAttribute('id');
        nav.removeAttribute('style');
        nav.removeAttribute('title');
        nav.classList.remove('active', 'disabled');
        nav.dataset.tab = 'music';
        const label = nav.querySelector('[data-i18n]');
        if (label) {
            label.setAttribute('data-i18n', 'nav.music');
            label.textContent = this.t('nav.music', 'Música');
        } else {
            nav.textContent = this.t('nav.music', 'Música');
        }
        const img = nav.querySelector('img');
        if (img) {
            img.src = '../assets/icons/tab_music.png';
            img.alt = 'Music';
        }
        ref.insertAdjacentElement('afterend', nav);
    }

    buildPane() {
        const ref = document.getElementById('pane-lighting');
        if (!ref || document.getElementById('pane-music')) return;
        const pane = document.createElement('div');
        pane.className = 'tab-pane';
        pane.id = 'pane-music';
        pane.innerHTML = `
          <div class="card">
            <div class="card-title">🎵 <span data-i18n="music.title">Modo música</span></div>
            <p data-i18n="music.help" style="font-size:13px;line-height:1.5;opacity:.85;margin:0 0 14px"></p>
            <div class="controls-grid">
              <div class="form-group">
                <label data-i18n="music.style">Estilo del efecto</label>
                <select id="music-style"></select>
              </div>
              <div class="form-group">
                <label data-i18n="music.bg_mode">Fondo</label>
                <select id="music-bg"></select>
              </div>
              <div class="form-group">
                <label><span data-i18n="music.bg_bright">Brillo del fondo</span>: <span class="slider-val" id="music-bright-val">100</span></label>
                <div class="slider-group"><input type="range" id="music-bright" min="0" max="100" value="100"></div>
              </div>
              <div class="form-group">
                <label><span data-i18n="music.amplitude">Amplitud (cuánto reacciona)</span>: <span class="slider-val" id="music-amp-val">1.0</span></label>
                <div class="slider-group"><input type="range" id="music-amp" min="10" max="400" value="100"></div>
              </div>
              <div class="form-group">
                <label data-i18n="music.source">Fuente de audio</label>
                <select id="music-source">
                  <option value="monitor" data-i18n="music.source_monitor">Lo que suena en el equipo</option>
                  <option value="mic" data-i18n="music.source_mic">Micrófono</option>
                </select>
              </div>
            </div>
            <div class="btn-row" style="margin-top:18px;align-items:center">
              <button class="btn" id="btn-music-toggle" data-i18n="music.start">Activar</button>
              <span id="music-status" style="margin-left:12px;font-size:13px;opacity:.85"></span>
            </div>
          </div>`;
        ref.insertAdjacentElement('afterend', pane);

        this.styleSel = pane.querySelector('#music-style');
        this.bgSel = pane.querySelector('#music-bg');
        this.brightInput = pane.querySelector('#music-bright');
        this.brightVal = pane.querySelector('#music-bright-val');
        this.ampInput = pane.querySelector('#music-amp');
        this.ampVal = pane.querySelector('#music-amp-val');
        this.sourceSel = pane.querySelector('#music-source');
        this.btn = pane.querySelector('#btn-music-toggle');
        this.statusEl = pane.querySelector('#music-status');

        // Los nombres exactos de cada estilo y fondo se confirman probando en el teclado.
        const styleWord = this.t('music.style_word', 'Estilo');
        const bgWord = this.t('music.bg_word', 'Fondo');
        for (let n = 1; n <= 7; n++) this.styleSel.add(new Option(`${styleWord} ${n}`, String(n)));
        for (let n = 0; n <= 8; n++) this.bgSel.add(new Option(`${bgWord} ${n}`, String(n)));
        this.styleSel.value = '2';
        this.bgSel.value = '0';

        this.styleSel.addEventListener('change', () => this.schedulePush());
        this.bgSel.addEventListener('change', () => this.schedulePush());
        this.brightInput.addEventListener('input', () => {
            this.brightVal.textContent = this.brightInput.value;
            this.schedulePush();
        });
        this.ampInput.addEventListener('input', () => {
            this.ampVal.textContent = (this.ampInput.value / 100).toFixed(1);
            this.schedulePush();
        });
        this.sourceSel.addEventListener('change', async () => {
            if (!this.running) return;
            await this.stop();   // cambiar la fuente exige reiniciar la captura
            await this.start();
        });
        this.btn.addEventListener('click', () => (this.running ? this.stop() : this.start()));
        this.renderStatus();
        if (window.i18n) window.i18n.applyTranslations(pane);
    }

    payload() {
        return {
            fore: parseInt(this.styleSel.value, 10),
            back: parseInt(this.bgSel.value, 10),
            bright: parseInt(this.brightInput.value, 10),
            gain: parseInt(this.ampInput.value, 10) / 100,
            source: this.sourceSel.value,
        };
    }

    // Ajustes en vivo: solo cambian lo que el motor manda; no se guarda nada en el teclado.
    schedulePush() {
        if (!this.running) return;
        clearTimeout(this.pushTimer);
        this.pushTimer = setTimeout(async () => {
            try {
                await window.hostBridge.invoke('music_start', this.payload());
            } catch (e) {
                this.log(`[ERROR] ${e.message}`, 'error');
            }
        }, 120);
    }

    async start() {
        try {
            await window.hostBridge.invoke('music_start', this.payload());
            this.setRunning(true);
            this.log('[OK] ' + this.t('music.started', 'Modo música activado'));
        } catch (e) {
            this.setRunning(false);
            this.log(`[ERROR] ${e.message}`, 'error');
            this.renderStatus(e.message);
        }
    }

    async stop() {
        try {
            await window.hostBridge.invoke('music_stop', {});
        } catch (e) {
            this.log(`[ERROR] ${e.message}`, 'error');
        }
        this.setRunning(false);
    }

    setRunning(on) {
        this.running = on;
        this.btn.setAttribute('data-i18n', on ? 'music.stop' : 'music.start');
        this.btn.textContent = on ? this.t('music.stop', 'Detener') : this.t('music.start', 'Activar');
        this.renderStatus();
        clearInterval(this.poll);
        this.poll = null;
        if (on) {
            // Esta consulta tambien avisa al servidor de que la página sigue abierta.
            this.poll = setInterval(async () => {
                try {
                    const st = await window.hostBridge.invoke('music_status', {});
                    if (!st.running) {
                        this.setRunning(false);
                        if (st.error) {
                            this.log(`[ERROR] ${st.error}`, 'error');
                            this.renderStatus(st.error);
                        }
                    }
                } catch (e) { /* se reintenta en el siguiente ciclo */ }
            }, 4000);
        }
    }

    renderStatus(error) {
        if (!this.statusEl) return;
        if (error) {
            this.statusEl.textContent = error;
        } else {
            this.statusEl.textContent = this.running
                ? '● ' + this.t('music.running', 'Activo')
                : '○ ' + this.t('music.stopped', 'Detenido');
        }
    }
}

window.TabMusic = TabMusic;
