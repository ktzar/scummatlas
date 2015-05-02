function setZplanePerspective(angle) {
    var distance = 40;
    angle *= -1;
    $('.zplane').each(function (i, el) {
        zOffset = i * distance;
        if (i === 0) {
            console.log(0.01 * angle)
            $(el).css('opacity', 1 - Math.abs(0.01 * angle));
        }
        $(el).css(
            'transform',
            'perspective(600px) rotateY(' + angle + 'deg) translateZ(' + zOffset + 'px)'
        );
    });
}

$(function () {
    $('.toggle-svg-bg').click(function() {
        $('.room-boxes').toggleClass('room-boxes__nobg');
    });

    if ($('.zplane').length > 0) {
        setZplanePerspective(0);
        $('#zPlane-rotation').on('input', function () {
            setZplanePerspective(parseInt($(this).val(), 10));
        });
    }
});
