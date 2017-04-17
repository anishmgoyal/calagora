(function( $ )
{

  window.doOfferDelete = function(id)
  {
    window.deleteOffer(id, "Revoke", function()
    {
      var elem = document.getElementById("offer-stub-" + id);
      if(elem)
      {
        elem.parentNode.removeChild(elem);
      }
    });
  };

})( jQuery );
