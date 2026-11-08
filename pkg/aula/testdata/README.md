# Paquetes golden de macros 2.4G

Cada archivo concatena paquetes de 33 bytes (report ID, 31 bytes de payload y checksum) extraídos sin modificar de las capturas USBPcap del programa oficial de AULA. Los `.pcapng` originales permanecen fuera del repositorio.

- `macro-one.bin`: `todo.pcapng`, desde offset decimal 11042; 16 trozos y cierre `09 04 10`. Una macro con cuatro eventos: W abajo, pausa 10 ms, W arriba, pausa 10 ms.
- `macro-three.bin`: `macros.pcapng`, desde offset decimal 56198; 22 trozos y cierre `09 08 16`. Tres macros con 19, 12 y 12 eventos; offsets de tabla 400, 484 y 540.

Se tomó el primer paquete de cada índice de trozo (escritura al dongle). Cada captura también incluye una respuesta con la misma cabecera y datos a cero.
