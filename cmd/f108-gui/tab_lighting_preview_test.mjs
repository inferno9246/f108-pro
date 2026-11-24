import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {test} from 'node:test';
import vm from 'node:vm';

test('los presets preparan controles sin enviar paquetes', () => {
    let sends = 0;
    const elements = Object.fromEntries(['lighting-color-hex', 'val-brightness', 'val-speed']
        .map(id => [id, {value: '', textContent: ''}]));
    const window = {hostBridge: {invoke: () => { sends++; }}};
    const document = {getElementById: id => elements[id]};
    const source = readFileSync(new URL('./web/js/tab_lighting.js', import.meta.url), 'utf8');
    vm.runInNewContext(source, {window, document});

    const app = {
        lightingModes: ['Off', 'Static', 'SingleOn', 'SingleOff', 'Glittering', 'Falling',
            'Colourful', 'Breath', 'Spectrum', 'Outward', 'Scrolling', 'Rolling'],
        hexToRgb: color => ({r: parseInt(color.slice(1, 3), 16),
            g: parseInt(color.slice(3, 5), 16), b: parseInt(color.slice(5, 7), 16)}),
        composer: {setCommand: () => {}},
    };
    const tab = new window.TabLighting(app);
    tab.modeSelect = {value: 'Rolling'};
    tab.colorPicker = {value: '#ff0000'};
    tab.sliderBrightness = {value: '5'};
    tab.sliderSpeed = {value: '3'};
    tab.rainbowCheck = {checked: true, disabled: false};
    tab.dirSelect = {value: '0'};
    tab.dirGroup = {style: {display: ''}};

    tab.previewPreset('warm');
    assert.equal(tab.modeSelect.value, 'Static');
    assert.equal(tab.colorPicker.value, '#ff9b5c');
    assert.equal(tab.sliderBrightness.value, '2');
    assert.equal(tab.sliderSpeed.value, '1');
    assert.equal(tab.rainbowCheck.checked, false);
    assert.equal(sends, 0);

    tab.previewPreset('wave');
    assert.equal(tab.modeSelect.value, 'Rolling');
    assert.equal(tab.rainbowCheck.checked, true);
    assert.equal(tab.sliderBrightness.value, '2');
    assert.equal(tab.sliderSpeed.value, '1');
    assert.equal(sends, 0);
});
