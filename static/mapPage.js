$(function(){ // on dom ready

// photos from flickr with creative commons license
  
var cy = cytoscape({
  container: document.getElementById('viewport'),
  
  style: cytoscape.stylesheet()
    .selector('node')
      .css({
        'height': 80,
        'width': 80,
        'background-fit': 'cover',
        'border-color': '#000',
        'border-width': 3,
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
      })
    .selector('#bird')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      })
    .selector('#cat')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      })
    .selector('#ladybug')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      })
  .selector('#aphid')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      })
  .selector('#rose')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      })
  .selector('#grasshopper')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      })
  .selector('#plant')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      })
  .selector('#wheat')
      .css({
        'background-image': 'img_bg/room08_bg.png'
      }),
  
  elements: {
    nodes: [
      { data: { id: 'cat' } },
      { data: { id: 'bird' } },
      { data: { id: 'ladybug' } },
      { data: { id: 'aphid' } },
      { data: { id: 'rose' } },
      { data: { id: 'grasshopper' } },
      { data: { id: 'plant' } },
      { data: { id: 'wheat' } }
    ],
    edges: [
      { data: { source: 'cat', target: 'bird' } },
      { data: { source: 'bird', target: 'ladybug' } },
      { data: { source: 'bird', target: 'grasshopper' } },
      { data: { source: 'grasshopper', target: 'plant' } },
      { data: { source: 'grasshopper', target: 'wheat' } },
      { data: { source: 'ladybug', target: 'aphid' } },
      { data: { source: 'aphid', target: 'rose' } }
    ]
  },
  
  layout: {
    name: 'breadthfirst',
    directed: true,
    padding: 10
  }
}); // cy init
  
/*
cy.on('tap', 'node', function(){
  var nodes = this;
  var tapped = nodes;
  var food = [];
  
  nodes.addClass('eater');
  
  for(;;){
    var connectedEdges = nodes.connectedEdges(function(){
      return !this.target().anySame( nodes );
    });
    
    var connectedNodes = connectedEdges.targets();
    
    Array.prototype.push.apply( food, connectedNodes );
    
    nodes = connectedNodes;
    
    if( nodes.empty() ){ break; }
  }
        
  var delay = 0;
  var duration = 500;
  for( var i = food.length - 1; i >= 0; i-- ){ (function(){
    var thisFood = food[i];
    var eater = thisFood.connectedEdges(function(){
      return this.target().same(thisFood);
    }).source();
            
    thisFood.delay( delay, function(){
      eater.addClass('eating');
    } ).animate({
      position: eater.position(),
      css: {
        'width': 10,
        'height': 10,
        'border-width': 0,
        'opacity': 0
      }
    }, {
      duration: duration,
      complete: function(){
        thisFood.remove();
      }
    });
    
    delay += duration;
  })(); } // for
  
}); // on tap
*/

}); // on dom ready
