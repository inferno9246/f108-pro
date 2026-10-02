// Oculta las imágenes que no se pudieron cargar (en lugar del icono roto).
document.addEventListener('error', function (e) {
    var t = e.target;
    if (t && t.tagName === 'IMG') { t.style.visibility = 'hidden'; }
}, true);
