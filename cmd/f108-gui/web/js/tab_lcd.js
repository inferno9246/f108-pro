/**
 * tab_lcd.js - TFT LCD Display (240x135) & Real-Time Clock Controller.
 */

class TabLcd {
    constructor(app) {
        this.app = app;
        this.selectedImagePath = null;
    }

    init() {
        this.syncClockBtn = document.getElementById('btn-lcd-sync-clock');
        this.selectFileBtn = document.getElementById('btn-lcd-select-file');
        this.flashBtn = document.getElementById('btn-lcd-flash');

        this.syncClockBtn.addEventListener('click', async () => {
            this.app.composer.setCommand('f108-pro clock');
            await this.app.composer.runCurrentCommand();
        });

        this.selectFileBtn.addEventListener('click', () => this.selectLcdFile());
        this.flashBtn.addEventListener('click', () => this.flashLcd());
    }

    async selectLcdFile() {
        try {
            const res = await window.hostBridge.invoke('open_file_dialog', {
                title: 'Select Image or Animated GIF for TFT LCD (240x135)',
                filter: 'image'
            });
            if (res && res.path) {
                this.selectedImagePath = res.path;
                this.app.composer.appendLog(`[LCD] Selected image: ${res.path}\n`, 'info');
                this.selectFileBtn.textContent = res.path.split('/').pop();
                this.app.composer.setCommand(`f108-pro lcd --preview ${res.path}`);
            }
        } catch (e) {
            console.error('File selection cancelled:', e);
        }
    }

    async flashLcd() {
        if (!this.selectedImagePath) {
            this.app.composer.appendLog('[WARN] Please select an image or GIF file first.\n', 'stderr');
            return;
        }

        try {
            this.flashBtn.disabled = true;
            this.flashBtn.textContent = 'Processing & Flashing...';
            this.app.composer.appendLog(`[LCD] Converting '${this.selectedImagePath}' to 240x135 RGB565...\n`, 'info');

            const res = await window.hostBridge.invoke('convert_and_flash_lcd', {
                image_path: this.selectedImagePath
            });

            if (res && res.success) {
                this.app.composer.appendLog(`[LCD] Flash successful! Frames uploaded: ${res.frames || 1}\n`, 'info');
            }
        } catch (err) {
            this.app.composer.appendLog(`[LCD ERROR] ${err.message}\n`, 'stderr');
        } finally {
            this.flashBtn.disabled = false;
            this.flashBtn.textContent = 'Upload and Flash to LCD';
        }
    }
}

window.TabLcd = TabLcd;
