$(function() { 

    var mapLayout = {
        name: 'breadthfirst',
        fit: true,
        directed: false,
        padding: 30,
        circle: false,
        spacingFactor: 1.75,
        boundingBox: undefined,
        avoidOverlap: true,
        roots: undefined,
        maximalAdjustments: 0,
        animate: false,
        animationDuration: 500,
        ready: undefined,
        stop: undefined
    };

    var style = cytoscape.stylesheet()
        .selector('node')
        .css({
            'height': 120,
            'width': 120,
            'background-fit': 'cover',
            'border-color': '#000',
            'border-width': 2,
            'border-opacity': 0.5
        })
        .selector('.eating')
        .css({
            'border-color': 'red'
        })
        .selector('.eater')
        .css({
            'border-width': 9
        })
        .selector('edge')
        .css({
            'width': 6,
            'target-arrow-shape': 'triangle',
            'line-color': '#ffaaaa',
            'target-arrow-color': '#ffaaaa'
        });
        for (var n in nodes) {
            node = nodes[n];
            nodeStyle = {
                'color': 'white',
                'text-outline-width': 2,
                'text-valign': 'center',
                'text-outline-color': '#888',
                'background-image': 'img_bg/' + node.data.id + '_bg.png',
                'content': ''+ node.data.name.toUpperCase() +''
            };
            style.selector('#' + node.data.id).css(nodeStyle);
        }

    var cy = cytoscape({
      container: document.getElementById('viewport'),
      style: style,
      layout: mapLayout,
      elements: {
        'nodes': nodes,
        'edges': edges
      },
    }); 

    cy.on('tap', 'node', function() {
        try { // your browser may block popups
            window.open(this.data('href'));
        } catch (e) { // fall back on url change
            window.location.href = this.data('href');
        }
    });
});
