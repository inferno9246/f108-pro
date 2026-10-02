/**
 * tab_remap.js - Physical key remapping over the 2.4G dongle.
 *
 * Derived from f108-pro-gui (MIT, see THIRD_PARTY.md). Instead of building CLI
 * strings and executing them, this version sends a JSON action to the Go server.
 */

class TabRemap {
    constructor(app) {
        this.app = app;
        this.layer = 'normal'; // 'normal' | 'fn'
        this.actionType = 'key'; // 'key' | 'combo' | 'consumer' | 'mouse'
        this.selectedKey = null;
        this.keyNames = [];
        this.nameByIndex = new Map();
        this.modInputs = {};
        this.modifiers = [
            { id: 'ctrl', name: 'Ctrl' }, { id: 'shift', name: 'Shift' },
            { id: 'alt', name: 'Alt' }, { id: 'win', name: 'Win' },
        ];
        this.consumerActions = [
            { id: 'volup', name: 'Subir volumen' }, { id: 'voldown', name: 'Bajar volumen' },
            { id: 'mute', name: 'Silenciar' }, { id: 'play', name: 'Reproducir / Pausa' },
            { id: 'stop', name: 'Detener' }, { id: 'next', name: 'Pista siguiente' },
            { id: 'prev', name: 'Pista anterior' },
        ];
        this.mouseActions = [
            { id: 'lclick', name: 'Clic izquierdo' }, { id: 'rclick', name: 'Clic derecho' },
            { id: 'mclick', name: 'Clic central' }, { id: 'scrollup', name: 'Rueda arriba' },
            { id: 'scrolldn', name: 'Rueda abajo' },
        ];
    }

    async init() {
        this.layerSelect = document.getElementById('remap-layer-select');
        this.actionTypeSelect = document.getElementById('remap-action-type');
        this.targetKeySelect = document.getElementById('remap-target-key-select');
        this.keyNameInput = document.getElementById('remap-selected-key-name');
        this.applyBtn = document.getElementById('btn-apply-remap');
        this.resetKeyBtn = document.getElementById('btn-reset-key');

        // Funciones aún sin soporte en esta versión.
        ['btn-export-yaml', 'btn-import-yaml'].forEach(id => {
            const b = document.getElementById(id);
            if (b) b.style.display = 'none';
        });
        const disabledOpt = this.actionTypeSelect.querySelector('option[value="disabled"]');
        if (disabledOpt) disabledOpt.remove();

        this.buildModifierBox();
        this.layerSelect.addEventListener('change', (e) => {
            this.layer = e.target.value === 'fn' ? 'fn' : 'normal';
            this.refreshMarks();
            this.updateCommand();
        });
        this.actionTypeSelect.addEventListener('change', (e) => {
            this.actionType = e.target.value;
            this.populateTargetKeys();
            this.updateCommand();
        });
        this.targetKeySelect.addEventListener('change', () => this.updateCommand());
        this.applyBtn.addEventListener('click', () => this.applyRemap());
        this.resetKeyBtn.addEventListener('click', () => this.resetKey());

        await this.loadKeyNames();
        this.bindResetAll();
        await this.loadState();
        this.populateTargetKeys();
    }

    log(msg, type = 'info') {
        if (this.app.composer && this.app.composer.appendLog) {
            this.app.composer.appendLog(msg + '\n', type);
        } else {
            console.log(msg);
        }
    }

    async loadKeyNames() {
        try {
            const res = await fetch('/api/keys', { headers: { 'X-Token': window.hostBridge.token } });
            if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
            const keys = await res.json();
            this.keyNames = keys.map(k => k.name);
            this.nameByIndex = new Map(keys.map(k => [k.index, k.name]));
        } catch (e) {
            this.log(`[ERROR] No se pudo cargar la lista de teclas: ${e.message}`, 'error');
        }
    }

    buildModifierBox() {
        const group = document.getElementById('remap-target-key-group');
        this.modBox = document.createElement('div');
        this.modBox.className = 'form-group';
        this.modBox.style.display = 'none';

        const title = document.createElement('label');
        title.textContent = 'Modifiers';
        this.modBox.appendChild(title);

        const row = document.createElement('div');
        row.style.display = 'flex';
        row.style.gap = '12px';
        this.modifiers.forEach(m => {
            const l = document.createElement('label');
            l.style.display = 'flex';
            l.style.alignItems = 'center';
            l.style.gap = '4px';
            const cb = document.createElement('input');
            cb.type = 'checkbox';
            cb.addEventListener('change', () => this.updateCommand());
            this.modInputs[m.id] = cb;
            l.appendChild(cb);
            l.appendChild(document.createTextNode(m.name));
            row.appendChild(l);
        });
        this.modBox.appendChild(row);
        group.insertAdjacentElement('afterend', this.modBox);
    }

    populateTargetKeys() {
        const sel = this.targetKeySelect;
        sel.innerHTML = '';
        let items;
        if (this.actionType === 'consumer') {
            items = this.consumerActions.map(a => [a.id, a.name]);
        } else if (this.actionType === 'mouse') {
            items = this.mouseActions.map(a => [a.id, a.name]);
        } else {
            items = this.keyNames.map(n => [n, n]);
        }
        items.forEach(([value, text]) => {
            const opt = document.createElement('option');
            opt.value = value;
            opt.textContent = text;
            sel.appendChild(opt);
        });
        if (this.actionType === 'key' && this.keyNames.includes('lctrl')) sel.value = 'lctrl';
        this.modBox.style.display = this.actionType === 'combo' ? '' : 'none';
    }

    // Destino en el formato del backend: "a", "media:play", "mouse:lclick", "combo:ctrl+c".
    targetSpec() {
        const v = this.targetKeySelect.value;
        if (!v) return '';
        switch (this.actionType) {
            case 'consumer': return `media:${v}`;
            case 'mouse': return `mouse:${v}`;
            case 'combo': {
                const mods = this.modifiers.filter(m => this.modInputs[m.id].checked).map(m => m.id);
                return mods.length ? `combo:${mods.join('+')}+${v}` : '';
            }
            default: return v;
        }
    }

    onKeySelected(keyObj) {
        this.selectedKey = keyObj || null;
        this.keyNameInput.value = keyObj ? `${keyObj.label} (${keyObj.id})` : '';
        this.updateCommand();
    }

    // Muestra el comando CLI equivalente (solo informativo; no se ejecuta).
    updateCommand() {
        const composer = this.app.composer;
        if (!composer || !this.targetKeySelect) return;
        if (!this.selectedKey) {
            composer.setCommand('f108-pro remap --help');
            return;
        }
        const fn = this.layer === 'fn' ? '--fn ' : '';
        const src = this.nameByIndex.get(this.selectedKey.key_index) || this.selectedKey.cli_name.toLowerCase();
        composer.setCommand(`f108-pro remap --wireless ${fn}${src} ${this.targetSpec() || '<target>'}`);
    }

    async applyRemap() {
        if (!this.selectedKey) return;
        const target = this.targetSpec();
        if (!target) {
            this.log('[ERROR] Elige un destino (y al menos un modificador si es un atajo).', 'error');
            return;
        }
        await this.send('remap_apply', { target }, `${this.selectedKey.label} -> ${target}`, true);
    }

    async resetKey() {
        if (!this.selectedKey) return;
        await this.send('remap_reset_key', {}, `${this.selectedKey.label} restaurada`, false);
    }

    async send(action, extra, okMsg, remapped) {
        this.applyBtn.disabled = true;
        this.resetKeyBtn.disabled = true;
        try {
            const res = await window.hostBridge.invoke(action, {
                transport: 'wireless',
                layer: this.layer,
                key_index: this.selectedKey.key_index,
                ...extra,
            });
            this.app.canvas.setKeyRemapped(this.selectedKey.id, remapped);
            this.noteState(remapped, extra.target);
            if (res && res.confirmed === false) {
                this.log(`[AVISO] ${okMsg} (${this.layer}): enviado, pero el teclado no lo confirmó. Pulsa la tecla para comprobar.`, 'info');
            } else {
                this.log(`[OK] ${okMsg} (${this.layer})`, 'info');
            }
        } catch (e) {
            this.log(`[ERROR] ${e.message}`, 'error');
        } finally {
            this.applyBtn.disabled = false;
            this.resetKeyBtn.disabled = false;
        }
    }
}


// --- Estado local y boton «Restaurar todas las modificadas» ---
TabRemap.prototype.refreshMarks = function () {
    const canvas = this.app.canvas;
    if (!canvas || !canvas.keysData) return;
    const marked = (this.state && this.state[this.layer]) || {};
    canvas.keysData.keys.forEach(k => canvas.setKeyRemapped(k.id, String(k.key_index) in marked));
};

TabRemap.prototype.loadState = async function () {
    try {
        const st = await window.hostBridge.invoke('remap_state');
        this.state = { normal: st.normal || {}, fn: st.fn || {} };
    } catch (e) {
        this.state = { normal: {}, fn: {} };
    }
    this.refreshMarks();
};

TabRemap.prototype.noteState = function (remapped, target) {
    this.state = this.state || { normal: {}, fn: {} };
    const idx = String(this.selectedKey.key_index);
    if (remapped) {
        this.state[this.layer][idx] = target;
    } else {
        delete this.state[this.layer][idx];
    }
};

TabRemap.prototype.bindResetAll = function () {
    const btn = document.getElementById('btn-reset-all');
    if (!btn) return;
    btn.addEventListener('click', async () => {
        const st = this.state || { normal: {}, fn: {} };
        const n = Object.keys(st.normal).length + Object.keys(st.fn).length;
        if (n === 0) {
            this.log('[INFO] No hay teclas modificadas desde esta interfaz.', 'info');
            return;
        }
        if (!confirm(`Se restauraran ${n} tecla(s) modificadas desde esta interfaz. ¿Continuar?`)) return;
        btn.disabled = true;
        try {
            const res = await window.hostBridge.invoke('remap_reset_all', {});
            this.state = { normal: {}, fn: {} };
            this.refreshMarks();
            this.log(`[OK] ${n} tecla(s) restauradas` + (res && res.unconfirmed ? ` (${res.unconfirmed} sin confirmar: comprueba pulsándolas)` : ''), 'info');
        } catch (e) {
            this.log(`[ERROR] ${e.message}`, 'error');
        } finally {
            btn.disabled = false;
        }
    });
};

window.TabRemap = TabRemap;
