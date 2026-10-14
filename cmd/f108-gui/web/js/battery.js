/**
 * battery.js - Indicador de bateria del teclado (dongle 2.4G).
 * Pregunta al servidor cada minuto; el nivel es el byte 3 de la respuesta al comando 20 01.
 */
(function () {
    function mount() {
        const ref = document.getElementById('status-text');
        if (!ref) return null;
        let el = document.getElementById('battery-status');
        if (!el) {
            el = document.createElement('span');
            el.id = 'battery-status';
            el.style.cssText = 'margin-left:10px;font-size:12px;opacity:.85';
            ref.insertAdjacentElement('afterend', el);
        }
        return el;
    }

    async function poll() {
        const el = mount();
        if (!el || !window.hostBridge) return;
        try {
            const st = await window.hostBridge.invoke('get_device_status');
            if (!st || !st.dongle_connected) {
                el.textContent = '';
                return;
            }
            const b = await window.hostBridge.invoke('battery');
            el.textContent = '🔋 ' + b.percent + ' %';
        } catch (e) {
            el.textContent = '';
        }
    }

    window.f108BatteryPoll = poll;
    window.addEventListener('DOMContentLoaded', () => {
        setTimeout(poll, 1500);
        setInterval(poll, 60000);
    });
})();
