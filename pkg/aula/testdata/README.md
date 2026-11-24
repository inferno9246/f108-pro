# Paquetes golden de macros 2.4G

Cada archivo concatena paquetes de 33 bytes (report ID, 31 bytes de payload y checksum) extraídos sin modificar de las capturas USBPcap del programa oficial de AULA. Los `.pcapng` originales permanecen fuera del repositorio.

- `macro-one.bin`: `todo.pcapng`, desde offset decimal 11042; 16 trozos y cierre `09 04 10`. Una macro con cuatro eventos: W abajo, pausa 10 ms, W arriba, pausa 10 ms.
- `macro-three.bin`: `macros.pcapng`, desde offset decimal 56198; 22 trozos y cierre `09 08 16`. Tres macros con 19, 12 y 12 eventos; offsets de tabla 400, 484 y 540.

Se tomó el primer paquete de cada índice de trozo (escritura al dongle). Cada captura también incluye una respuesta con la misma cabecera y datos a cero.

- `settings-seventeen.bin`: `todo.pcapng`, 17 envíos `07 10` de 33 bytes en los offsets decimales 3330, 3634, 3938, 4242, 4850, 5154, 5458, 5762, 6066, 6370, 6978, 7282, 7586, 7890, 8194, 8498 y 8802. Incluye todos los niveles de respuesta 1-5, reposo 0-3 y las tres casillas activadas por separado. Se excluyeron sus ACK.
- `light-screen-effect-one.bin`: `todo.pcapng`, paquete `05 10` en offset decimal 9106; modo 11, multicolor, brillo 5 y velocidad 3. Coincide con `default_mode=11` del XML oficial; el usuario confirmó visualmente en hardware que la onda de izquierda a derecha corresponde al «efecto 1» de la pantalla.
- `clock-one.bin`: `todo.pcapng`, paquete `0C 10` en offset decimal 24550; 6 de octubre de 2026, 20:31:31, martes (`weekday=2`). El test compara los 33 bytes, incluido el checksum, con `SyncClock` usando esa fecha.
- `battery-query.bin` y `battery-response.bin`: `todo.pcapng`, offsets decimales 3026 y 3178. Consulta `20 01` y respuesta con byte de batería `0x5B` (91 %); el test compara la consulta completa y analiza una respuesta real capturada.
