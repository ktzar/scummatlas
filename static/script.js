function setZplanePerspective(angle) {
    var distance = 40, zOffset, perspective;
    $('.zplane').each(function (i, el) {
        zOffset = i * distance;
        perspective = 900;
        if (i === 0) {
            $(el).css('opacity', 1 - Math.abs(0.01 * angle));
        }
        $(el).css(
            'transform',
            [
                'perspective(' + perspective + 'px)',
                'rotateY(' + angle + 'deg)',
                'translateZ(' + zOffset + 'px)'
            ].join(" ")
        );
    });
}

$(function () {
    $('.toggle-svg-bg').click(function() {
        $('.room-boxes').toggleClass('room-boxes__nobg');
    });

    $('.zplanes-container').on('wheel', function (evt) {
        var control = $('#zPlane-rotation'), newVal;
        evt.preventDefault();
        newVal = parseInt(control.val(), 10) + evt.wheelDelta;
        control.val(newVal);
        setZplanePerspective(control.val());
    });

    if ($('.zplane').length > 0) {
        setZplanePerspective(0);
        $('#zPlane-rotation').on('input', function () {
            setZplanePerspective(parseInt($(this).val(), 10));
        });
    }
});
