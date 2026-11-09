# Pruebas 2.4G y retorno al estado anterior

Antes de una prueba que escriba al teclado:

1. Copiar `~/.config/f108-pro/gui-*.json` a un directorio privado con fecha y
   permisos `0700`. Esos archivos guardan lo enviado por la GUI; **no son una
   lectura del estado completo del teclado**.
2. Anotar el efecto visible, ajustes, macros y teclas modificadas. Preparar la
   acción de retorno usando solo paquetes ya capturados y probados con golden
   tests. Si no se puede reconstruir el estado anterior, no ejecutar la prueba.
3. Enviar solo el paquete conocido. Comprobar ACK y el comportamiento visible.
4. Restaurar el estado anotado y comprobar otra vez ACK y comportamiento. No
   probar carga de imágenes LCD, firmware ni reset de fábrica.

Estado confirmado el 8 de noviembre de 2026: efecto 1 de la pantalla = modo
2.4G `11` (multicolor, brillo 5, velocidad 3), reposo nivel 2 (5 minutos),
respuesta nivel 1, macro W en la ranura 0 e Insert asignada a `macro:0`. El
efecto fue confirmado visualmente y los envíos de efecto y ajustes recibieron
ACK. La persistencia tras apagar y el estado de otras teclas no se han leído.

Para volver a este estado desde la GUI: **Inicio → Efecto 1 del teclado**,
**Ajustes → 5 minutos / nivel 1 → Aplicar**, y, solo si una prueba cambió
macros/remapeos, **Macros → enviar W capturada → asignar macro 0 a Insert**.
Cada paso vuelve a enviar paquetes capturados; verificar el resultado en el
teclado después de aplicarlo.
