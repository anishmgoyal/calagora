(function()
{

  window.lightPager = function(ndTarget, query, min, currentPage, max)
  {
    var numForStage = [3, 4, 2];
    var startForStage = [min,
      currentPage-Math.floor(numForStage[1]/2),
      max-numForStage[2]+1];

    if(min != max)
    {
      // Start rendering for small screens and windows
      if(currentPage - 1 > min)
      {
        var ndSkipFirst = document.createElement("div");
        ndSkipFirst.className = "pagingButton pagingButtonArrow if-small";
        ndSkipFirst.innerHTML = "&#10094;&#10094;";
        ndSkipFirst.onclick = function()
        {
          window.location.href = "/search/?q="+
            encodeURIComponent(query)+"&page="+
            (min);
        }
        ndTarget.appendChild(ndSkipFirst);
      }
      if(currentPage > min)
      {
        var ndBackOne = document.createElement("div");
        ndBackOne.className = "pagingButton pagingButtonArrow if-small";
        ndBackOne.innerHTML = "&#10094;";
        ndBackOne.onclick = function()
        {
          window.location.href = "/search/?q="+
            encodeURIComponent(query)+"&page="+
            (currentPage-1);
        };
        ndTarget.appendChild(ndBackOne);
      }
      var ndCurrentPage = document.createElement("div");
      ndCurrentPage.className = "pagingButton pagingButtonArrow if-small"+
        " pagingButtonCurrent";
      ndCurrentPage.appendChild(document.createTextNode(
        currentPage + " of " + max
      ));
      ndTarget.appendChild(ndCurrentPage);
      if(currentPage < max)
      {
        var ndForwardOne = document.createElement("div");
        ndForwardOne.className = "pagingButton pagingButtonArrow if-small";
        ndForwardOne.innerHTML = "&#10095;";
        ndForwardOne.onclick = function()
        {
          window.location.href="/search/?q="+
            encodeURIComponent(query)+"&page="+
            (currentPage+1);
        };
        ndTarget.appendChild(ndForwardOne);
      }
      if(currentPage + 1 < max)
      {
        var ndSkipLast = document.createElement("div");
        ndSkipLast.className = "pagingButton pagingButtonArrow if-small";
        ndSkipLast.innerHTML = "&#10095;&#10095;";
        ndSkipLast.onclick = function()
        {
          window.location.href="/search/?q="+
            encodeURIComponent(query)+"&page="+
            (max);
        };
        ndTarget.appendChild(ndSkipLast);
      }

      // Start rendering for larger screens
      if(currentPage > min)
      {
        ndBackOne = document.createElement("div");
        ndBackOne.className = "pagingButton pagingButtonArrow unless-small";
        ndBackOne.innerHTML = "&#10094;";
        ndBackOne.onclick = function()
        {
          window.location.href = "/search/?q="+
            encodeURIComponent(query)+"&page="+
            (currentPage-1);
        };
        ndTarget.appendChild(ndBackOne);
      }
      var prevEnd = min - 1;
      for(var i = 0; i < numForStage.length; i++)
      {
        var currentStart = startForStage[i];
        if(currentStart > max)
        {
          break;
        }

        if(currentStart > prevEnd + 1)
        {
          var ndEllipsis = document.createElement("span");
          ndEllipsis.className = "unless-small";
          ndEllipsis.appendChild(document.createTextNode("... "));
          ndTarget.appendChild(ndEllipsis);
        }

        var current = currentStart;
        var numWritten = 0;
        while(numWritten < numForStage[i] && current <= max)
        {
          var extraClass = (current == currentPage)? "pagingButtonCurrent" : "";

          if(current > prevEnd)
          {
            prevEnd = current;
            var ndPage = document.createElement("div");
            ndPage.className = "pagingButton " + extraClass +
              " unless-small";
            ndPage.appendChild(document.createTextNode(current));
            ndPage.onclick = function(current)
            {
              window.location.href = "/search/?q="+
                encodeURIComponent(query)+"&page="+
                (current);
            }.bind(window, current);
            ndTarget.appendChild(ndPage);
          }
          current++;
          numWritten++;
        }
      }
      if(currentPage < max)
      {
        ndForwardOne = document.createElement("div");
        ndForwardOne.className = "pagingButton pagingButtonArrow "+
          "unless-small";
        ndForwardOne.innerHTML = "&#10095;";
        ndForwardOne.onclick = function()
        {
          window.location.href = "/search/?q="+
            encodeURIComponent(query)+"&page="+
            (currentPage+1);
        };
        ndTarget.appendChild(ndForwardOne);
      }
    }
  }

})();
