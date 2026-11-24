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
ACK. El estado de otras teclas no se ha leído.

El 24 de noviembre de 2026, el usuario volvió a confirmar efecto 1 e Insert
escribiendo `w`. Tras apagar y encender en 2.4G, ambos resultados seguían
presentes. La GUI seguía mostrando reposo de 5 minutos y respuesta nivel 1,
pero son valores guardados localmente, no lecturas del teclado. Los JSON se
respaldaron en `~/.config/f108-pro/backups/20261124T185111Z/`; este respaldo
no contiene una lectura completa del teclado. Las tres casillas sin nombre
podrían haber cambiado desde noviembre, así que no se enviará `07 10` hasta
conocer sus valores actuales.

En la misma sesión, el usuario pasó a USB-C con FN+4 y comprobó que Insert
seguía escribiendo `w`; al volver a 2.4G con FN+~, también escribió `w`.
No se enviaron paquetes durante esos cambios de transporte.

Para volver a este estado desde la GUI: **Inicio → Efecto 1 del teclado**,
**Ajustes → 5 minutos / nivel 1 → Aplicar**, y, solo si una prueba cambió
macros/remapeos, **Macros → enviar W capturada → asignar macro 0 a Insert**.
Cada paso vuelve a enviar paquetes capturados; verificar el resultado en el
teclado después de aplicarlo.
