/**
 * keyboard_canvas.js - Interactive 108-Key Matrix Canvas Engine.
 *
 * Renders the 104 clickable keycap overlays over the official 800x300 3D layout.
 * Supports single key selection, per-key RGB coloring, and remap status indicators.
 */

class KeyboardCanvas {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.keysData = null;
        this.keyElements = new Map(); // keyId -> DOM button element
        this.selectedKey = null;
        this.onKeySelectedCallback = null;
        this.onScreenClickedCallback = null;
    }

    /**
     * Initializes and renders the keyboard canvas from JSON coordinates.
     */
    async init() {
        if (!this.container) return;

        try {
            const res = await fetch('../assets/keyboard/key_coordinates.json');
            this.keysData = await res.json();
            // Esta unidad no tiene Print/Scroll/Pause (ahí van la pantalla y la perilla).
            // (Print/Scroll/Pause se muestran bloqueadas, no se ocultan)
            this.render();
        } catch (err) {
            console.error('Failed to load key_coordinates.json:', err);
        }
    }

    /**
     * Builds DOM structure.
     */
    render() {
        this.container.innerHTML = '';
        this.keyElements.clear();

        // Stage container (800x300)
        const stage = document.createElement('div');
        stage.className = 'keyboard-stage';
        const probe = new Image();
        probe.onload = () => stage.classList.add('con-foto');
        probe.src = '/assets/keyboard/img_keyboard_layout.png';

        // Render each key
        this.keysData.keys.forEach(k => {
            const btn = document.createElement('button');
            btn.className = 'keycap';
            btn.dataset.id = k.id;
            btn.dataset.code = k.code;
            btn.dataset.keyIndex = k.key_index;
            btn.dataset.lightIndex = k.light_index;
            btn.title = `${k.label} (${k.id}) | Code: ${k.code} | Index: ${k.key_index}`;

            // Coordinates in pixels
            btn.style.left = `${k.rect.x}px`;
            btn.style.top = `${k.rect.y}px`;
            btn.style.width = `${k.rect.w}px`;
            btn.style.height = `${k.rect.h}px`;

            // Label span
            const label = document.createElement('span');
            label.className = 'key-label';
            label.textContent = k.label;
            btn.appendChild(label);

            // Click listener
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.selectKey(k.id);
            });

            stage.appendChild(btn);
            if (k.id === 'FN' || k.cli_name === 'FN' || [112, 113, 115].includes(k.key_index)) {
                btn.disabled = true;
                btn.classList.add('locked');
                btn.title = `${k.label}: no configurable`;
            }
            this.keyElements.set(k.id, btn);
        });

        // TFT Screen Hotspot
        const screenHotspot = document.createElement('div');
        screenHotspot.className = 'screen-hotspot';
        screenHotspot.title = 'TFT LCD Screen & Knob (240x135) - Click to configure';
        screenHotspot.innerHTML = '<span>TFT LCD / RTC</span>';
        screenHotspot.addEventListener('click', (e) => {
            e.stopPropagation();
            if (this.onScreenClickedCallback) this.onScreenClickedCallback();
        });
        stage.appendChild(screenHotspot);

        this.container.appendChild(stage);
    }

    /**
     * Selects a key by its ID (e.g. 'CAPSLOCK', 'ESC', 'Q').
     */
    selectKey(keyId) {
        // Deselect previous
        if (this.selectedKey && this.keyElements.has(this.selectedKey)) {
            this.keyElements.get(this.selectedKey).classList.remove('selected');
        }

        this.selectedKey = keyId;

        if (keyId && this.keyElements.has(keyId)) {
            const el = this.keyElements.get(keyId);
            el.classList.add('selected');
            const keyObj = this.keysData.keys.find(k => k.id === keyId);
            if (this.onKeySelectedCallback) {
                this.onKeySelectedCallback(keyObj, el);
            }
        } else {
            if (this.onKeySelectedCallback) {
                this.onKeySelectedCallback(null, null);
            }
        }
    }

    /**
     * Sets individual simulated LED backlighting on a keycap.
     */
    setKeyLed(keyId, hexColor) {
        if (this.keyElements.has(keyId)) {
            const el = this.keyElements.get(keyId);
            if (hexColor && hexColor !== 'transparent' && hexColor !== '#000000') {
                el.style.setProperty('--key-led-color', hexColor);
                el.classList.add('led-active');
            } else {
                el.style.removeProperty('--key-led-color');
                el.classList.remove('led-active');
            }
        }
    }

    /**
     * Clears all simulated LED backlighting on all keys.
     */
    clearAllLeds() {
        this.keyElements.forEach(el => {
            el.style.removeProperty('--key-led-color');
            el.classList.remove('led-active');
        });
    }

    /**
     * Sets a key remapped visual indicator dot.
     */
    setKeyRemapped(keyId, isRemapped) {
        if (this.keyElements.has(keyId)) {
            const el = this.keyElements.get(keyId);
            if (isRemapped) {
                el.classList.add('remapped');
            } else {
                el.classList.remove('remapped');
            }
        }
    }

    /**
     * Callback setters
     */
    onKeySelected(cb) {
        this.onKeySelectedCallback = cb;
    }

    onScreenClicked(cb) {
        this.onScreenClickedCallback = cb;
    }
}

window.KeyboardCanvas = KeyboardCanvas;
