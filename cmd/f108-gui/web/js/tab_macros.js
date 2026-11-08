/** Macros y asignaciones por el dongle 2.4G. */
class TabMacros {
    constructor(app) {
        this.app = app;
        this.macros = [];
        this.selected = -1;
        this.dirty = false;
        this.sent = false;
        this.recording = false;
        this.lastEventAt = null;
    }

    t(key, fallback) { return window.i18n.t(`macros.${key}`, fallback); }
    el(id) { return document.getElementById(`macro-${id}`); }
    log(message, type = 'info') { this.app.composer.appendLog(message + '\n', type); }

    async init() {
        this.select = this.el('select');
        this.name = this.el('name');
        this.events = this.el('events');
        this.bindMacro = this.el('bind-macro');
        this.populateKeys();
        this.el('add').addEventListener('click', () => this.addMacro());
        this.el('delete').addEventListener('click', () => this.removeMacro());
        this.el('up').addEventListener('click', () => this.moveMacro(-1));
        this.el('down').addEventListener('click', () => this.moveMacro(1));
        this.select.addEventListener('change', () => { this.stopRecord(); this.selected = Number(this.select.value); this.render(); });
        this.name.addEventListener('input', () => {
            if (this.selected < 0) return;
            this.macros[this.selected].name = this.name.value;
            this.dirty = true;
            this.select.options[this.selected].textContent = `${this.selected}: ${this.name.value || this.t('unnamed', 'Sin nombre')}`;
            this.renderSize();
        });
        this.el('event-type').addEventListener('change', () => this.renderEventFields());
        this.el('add-event').addEventListener('click', () => this.addEvent());
        this.el('record').addEventListener('click', () => this.toggleRecord());
        this.el('send').addEventListener('click', () => this.send());
        this.el('bind-mode').addEventListener('change', () => this.renderBindCount());
        this.el('bind').addEventListener('click', () => this.bind());
        document.addEventListener('keydown', (e) => this.capture(e, 'down'), true);
        document.addEventListener('keyup', (e) => this.capture(e, 'up'), true);
        window.addEventListener('localeChanged', () => this.render());
        try {
            const state = await window.hostBridge.invoke('macro_state');
            this.macros = Array.isArray(state.macros) ? state.macros : [];
            this.sent = this.macros.length > 0;
            this.selected = this.macros.length ? 0 : -1;
        } catch (e) { this.log(`[ERROR] ${e.message}`, 'error'); }
        this.render();
    }

    populateKeys() {
        const keys = this.app.tabRemap.keyNames.filter(k => k !== 'fn');
        for (const id of ['event-key', 'bind-key']) {
            const sel = this.el(id);
            sel.replaceChildren(...keys.filter(k => id !== 'bind-key' || !['printscreen', 'scrolllock', 'pause'].includes(k))
                .map(k => new Option(k, k)));
        }
        this.el('bind-key').value = 'insert';
    }

    addMacro() {
        if (this.macros.length >= 100) return this.log(`[ERROR] ${this.t('limit', 'Máximo 100 macros')}`, 'error');
        this.stopRecord();
        this.macros.push({name: `${this.t('default_name', 'Macro')} ${this.macros.length + 1}`, events: []});
        this.selected = this.macros.length - 1;
        this.dirty = true;
        this.render();
    }

    removeMacro() {
        if (this.selected < 0) return;
        if (this.macros.length === 1 && this.sent) {
            return this.log(`[ERROR] ${this.t('last_warning', 'El borrado total espera una captura oficial de la tabla vacía')}`, 'error');
        }
        this.stopRecord();
        this.macros.splice(this.selected, 1);
        this.selected = Math.min(this.selected, this.macros.length - 1);
        this.dirty = true;
        this.render();
    }

    moveMacro(direction) {
        const next = this.selected + direction;
        if (this.selected < 0 || next < 0 || next >= this.macros.length) return;
        this.stopRecord();
        [this.macros[this.selected], this.macros[next]] = [this.macros[next], this.macros[this.selected]];
        this.selected = next;
        this.dirty = true;
        this.render();
    }

    render() {
        const chosenBind = this.bindMacro.value;
        this.select.replaceChildren(...this.macros.map((m, i) => new Option(`${i}: ${m.name || this.t('unnamed', 'Sin nombre')}`, String(i))));
        this.bindMacro.replaceChildren(...this.macros.map((m, i) => new Option(`${i}: ${m.name || this.t('unnamed', 'Sin nombre')}`, String(i))));
        this.bindMacro.value = chosenBind && Number(chosenBind) < this.macros.length ? chosenBind : String(Math.max(0, this.selected));
        if (this.selected >= 0) this.select.value = String(this.selected);
        this.name.disabled = this.selected < 0;
        this.name.value = this.selected < 0 ? '' : this.macros[this.selected].name;
        this.el('delete').disabled = this.selected < 0;
        this.el('up').disabled = this.selected <= 0;
        this.el('down').disabled = this.selected < 0 || this.selected >= this.macros.length - 1;
        this.el('add-event').disabled = this.selected < 0;
        this.el('record').disabled = this.selected < 0;
        this.events.replaceChildren();
        if (this.selected >= 0) {
            this.macros[this.selected].events.forEach((event, i) => {
                const li = document.createElement('li');
                const label = document.createElement('code');
                label.textContent = event;
                const remove = document.createElement('button');
                remove.type = 'button';
                remove.className = 'btn btn-secondary';
                remove.textContent = '×';
                remove.setAttribute('aria-label', `${this.t('remove_event', 'Eliminar evento')} ${i + 1}`);
                remove.addEventListener('click', () => {
                    this.macros[this.selected].events.splice(i, 1);
                    this.dirty = true;
                    this.render();
                });
                li.append(label, remove);
                this.events.append(li);
            });
        }
        this.renderEventFields();
        this.renderBindCount();
        this.renderSize();
        this.el('bind').disabled = !this.sent || this.dirty || this.macros.length === 0;
    }

    renderEventFields() {
        const type = this.el('event-type').value;
        this.el('key-group').hidden = !['down', 'up', 'tap'].includes(type);
        this.el('ms-group').hidden = !['delay', 'tap'].includes(type);
        this.el('text-group').hidden = type !== 'text';
    }

    renderBindCount() { this.el('bind-count-group').hidden = this.el('bind-mode').value !== 'times'; }

    addEvent() {
        if (this.selected < 0) return;
        const type = this.el('event-type').value;
        const key = this.el('event-key').value;
        const ms = Number(this.el('event-ms').value);
        let event;
        if (type === 'delay' || type === 'tap') {
            if (!Number.isInteger(ms) || ms < 10 || ms > 65535) return this.log(`[ERROR] ${this.t('invalid_ms', 'Pausa inválida (10–65535 ms)')}`, 'error');
        }
        if (type === 'text') {
            const value = this.el('event-text').value;
            if (!/^[a-z0-9]+$/.test(value)) return this.log(`[ERROR] ${this.t('invalid_text', 'Texto: solo a–z y 0–9')}`, 'error');
            event = `text:${value}`;
        } else if (type === 'delay') event = `delay:${ms}`;
        else if (type === 'tap') event = `tap:${key}:${ms}`;
        else event = `${type}:${key}`;
        this.macros[this.selected].events.push(event);
        this.dirty = true;
        this.render();
    }

    // KeyboardEvent.code es independiente de la distribución activa de Linux.
    keyFromCode(code) {
        if (/^Key[A-Z]$/.test(code)) return code.slice(3).toLowerCase();
        if (/^Digit[0-9]$/.test(code)) return code.slice(5);
        if (/^F([1-9]|1[0-2])$/.test(code)) return code.toLowerCase();
        if (/^Numpad[0-9]$/.test(code)) return `num${code.slice(6)}`;
        const names = {
            Escape: 'esc', Backquote: 'grave', Minus: 'minus', Equal: 'equal', Backspace: 'backspace',
            Tab: 'tab', BracketLeft: 'lbracket', BracketRight: 'rbracket', Backslash: 'backslash',
            CapsLock: 'capslock', Semicolon: 'semicolon', Quote: 'quote', Enter: 'enter',
            ShiftLeft: 'lshift', ShiftRight: 'rshift', ControlLeft: 'lctrl', ControlRight: 'rctrl',
            AltLeft: 'lalt', AltRight: 'ralt', MetaLeft: 'lwin', MetaRight: 'rwin',
            Comma: 'comma', Period: 'dot', Slash: 'slash', Space: 'space',
            Insert: 'insert', Delete: 'delete', Home: 'home', End: 'end', PageUp: 'pageup', PageDown: 'pagedown',
            ArrowUp: 'up', ArrowDown: 'down', ArrowLeft: 'left', ArrowRight: 'right',
            NumLock: 'numlock', NumpadDivide: 'numslash', NumpadMultiply: 'numstar',
            NumpadSubtract: 'numminus', NumpadAdd: 'numplus', NumpadEnter: 'numenter', NumpadDecimal: 'numdot',
        };
        return names[code] || null;
    }

    toggleRecord() {
        if (this.recording) return this.stopRecord();
        if (this.selected < 0) return;
        this.recording = true;
        this.lastEventAt = null;
        this.el('record').textContent = this.t('stop_record', 'Detener grabación');
        this.el('record-status').textContent = this.t('recording', 'Grabando…');
        this.el('record').focus();
    }

    stopRecord() {
        if (!this.recording) return;
        this.recording = false;
        this.lastEventAt = null;
        this.el('record').textContent = this.t('record', 'Grabar teclas');
        this.el('record-status').textContent = '';
    }

    capture(e, action) {
        if (!this.recording || this.app.activeTab !== 'macros') return;
        e.preventDefault();
        e.stopPropagation();
        if (e.repeat) return;
        const key = this.keyFromCode(e.code);
        if (!key || !this.app.tabRemap.keyNames.includes(key) || key === 'fn') return;
        const now = performance.now();
        const events = this.macros[this.selected].events;
        if (this.lastEventAt !== null) events.push(`delay:${Math.min(65535, Math.max(10, Math.round(now - this.lastEventAt)))}`);
        events.push(`${action}:${key}`);
        this.lastEventAt = now;
        this.dirty = true;
        this.render();
    }

    eventCount(event) {
        if (event.startsWith('tap:')) return 4;
        if (event.startsWith('text:')) return [...event.slice(5)].length * 4;
        return 1;
    }

    renderSize() {
        const count = this.macros.reduce((sum, m) => sum + m.events.reduce((n, e) => n + this.eventCount(e), 0), 0);
        const bytes = 400 + this.macros.length * 8 + count * 4;
        const limitOK = bytes <= 3584 && (this.macros.length + count) * 8 <= 3600;
        this.el('size').textContent = `${this.t('size', 'Tamaño de tabla')}: ${bytes}/3584 ${this.t('bytes', 'bytes')} · ${count} ${this.t('event_count', 'eventos')}${limitOK ? '' : ' · ' + this.t('too_large', 'Límite excedido')}`;
        this.el('size').classList.toggle('macro-error', !limitOK);
        this.el('send').disabled = !limitOK || this.macros.length === 0 || this.macros.some(m => m.events.length === 0);
        this.el('bind').disabled = !this.sent || this.dirty || this.macros.length === 0;
    }

    async send() {
        this.stopRecord();
        const button = this.el('send');
        button.disabled = true;
        try {
            await window.hostBridge.invoke('macro_send', {macros: this.macros});
            this.sent = true;
            this.dirty = false;
            this.log(`[OK] ${this.t('sent', 'Tabla de macros enviada')}`);
        } catch (e) { this.log(`[ERROR] ${e.message}`, 'error'); }
        this.renderSize();
    }

    async bind() {
        if (!this.sent || this.dirty) return;
        const index = Number(this.bindMacro.value);
        const key = this.el('bind-key').value;
        const keyIndex = this.app.tabRemap.keyNames.indexOf(key) >= 0
            ? [...this.app.tabRemap.nameByIndex].find(([, name]) => name === key)?.[0] : undefined;
        if (keyIndex === undefined || index < 0 || index >= this.macros.length) return;
        const mode = this.el('bind-mode').value;
        const count = Number(this.el('bind-count').value);
        if (mode === 'times' && (!Number.isInteger(count) || count < 1 || count > 255)) return this.log(`[ERROR] ${this.t('invalid_count', 'Repeticiones: 1–255')}`, 'error');
        let target = `macro:${index}`;
        if (mode === 'times') target += `:times:${count}`;
        else if (mode !== 'once') target += `:${mode}`;
        try {
            const result = await window.hostBridge.invoke('remap_apply', {
                transport: 'wireless', layer: this.el('bind-layer').value, key_index: keyIndex, target,
            });
            this.log(`[${result.confirmed ? 'OK' : 'AVISO'}] ${key} → ${target}`);
            await this.app.tabRemap.loadState();
        } catch (e) { this.log(`[ERROR] ${e.message}`, 'error'); }
    }
}

window.TabMacros = TabMacros;
