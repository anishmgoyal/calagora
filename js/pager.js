(function( $ ){

  /**
   * Creates a new pager and initializes the parameters therein to provided parameters
   * addPagerFunctionality: if true, binds click handlers, wraps functionality in provided functions
   *                        such as hiding the load button, showing a load icon (if available), and
   *                        automatically removing the load more button if it seems we have loaded
   *                        the last page
   * button: the element which when clicked loads the next page
   * currentPage: the page which the pager will load first
   * itemList: optional, a property of data to use as the list of returned items
   * pageSize: how many items should be on each page
   * loadingIcon: an element to insert in the button's place as a loading indicator
   * loadPageFunction: an apiHandler call which loads a page
   *                   after static arguments, must take:
   *                   int page, int pageSize, fn successCallback(data), fn errorCallback(code)
   *                   page: the page to load
   *                   pageSize: how many items should be on each page
   *                   successCallback: takes the data returned by the api call on success
   *                   errorCallback: gets an HTTP error code if an error occurs
   * loadPageApi: the api object being used for the page load
   * loadPageParameters: an array of the first few arguments loadPageFunction accepts
   * loadPageSuccess: provided to the loadPageFunction, called on page load success. takes a data object as an argument.
   * loadPageError: provided to the loadPageFunction, called on page load error. takes an HTTP error code as an argument.
   * loadPageFinal: provided to the loadPageFunction. called on success if the number of returned items is less than the page size.
   *                only used IF addPagerFunctionality is set to true
   * useJqueryEvents: if true, all bound events will use jquery event handlers
   */
  var pager = function pager_cons(params) {

      if (this.constructor !== pager_cons) return new pager_cons(params);

      var settings = {
          addPagerFunctionality: true,
          button: null,
          currentPage: 0,
          itemList: "",
          pageSize: 50,
          listName: null,
          loadPageFunction: null,
          loadPageApi: window,
          loadPageParameters: [],
          loadPageSuccess: null,
          loadPageError: null,
          loadPageFinal: null,
          useJqueryEvents: false
      };

      for (var param in params) {
          if (params.hasOwnProperty(param)) {
              settings[param] = params[param];
          }
      }

      if (!settings.hasOwnProperty("loadingIcon")) {
          settings.loadingIcon = createLoadingIcon();
      }

      this.wrapFunctionality(settings);

      this.settings = settings;

      return this;

  };

  /**
   * Adds in-built functionality to the provided functions to reduce the amount of code
   * that has to be written for a paged view
   */
  pager.prototype.wrapFunctionality = function (settings) {

      // Add loading icon to next page functionality
      if (settings.loadingIcon != null) {
          var showLoadingIcon = function (settings) {
              if (settings.button != null) {
                  settings.button.parentNode.insertBefore(settings.loadingIcon, settings.button);
                  settings.button.style.display = "none";
              }
          }.bind(this, settings);

          this.nextPage = coupleFunctions(showLoadingIcon, this.nextPage);
      }

      // Add hide load icon functionality
      if (settings.loadingIcon != null) {
          var hideLoadingIcon = function (settings) {
              if (settings.button != null) {
                  settings.loadingIcon.parentNode.removeChild(settings.loadingIcon);
                  settings.button.style.display = "inline-block";
              }
          }.bind(this, settings);

          settings.loadPageSuccess = coupleFunctions(settings.loadPageSuccess, hideLoadingIcon);
      }

      // Automatically remove button on final
      var finalFN = function () {
          if (settings.button != null) {
              this.settings.button.parentNode.removeChild(this.settings.button);
          }
      }.bind(this);
      settings.loadPageFinal = coupleFunctions(settings.loadPageFinal, finalFN);

      // Wrap final functionality into the success function
      var successFN = (function (settings) {
          return function (data) {
              if (settings.listName != null)
              {
                data = data[settings.listName];
              }
              if(settings.itemList)
              {
                data = data[settings.itemList];
              }

              if (data.length < settings.pageSize) {
                  settings.loadPageFinal.call(window);
              }
          };
      })(settings);
      settings.loadPageSuccess = coupleFunctions(settings.loadPageSuccess, successFN);

      // Wrap click functionality
      if (settings.button != null) {
          var handler = this.nextPage.bind(this);
          if (settings.useJqueryEvents) {
              $(settings.button).click(handler);
          } else {
              if (settings.button.onclick != null) {
                  handler = coupleFunctions(handler, settings.button.onclick);
              }
              settings.button.onclick = handler;
          }
      }

  };

  /**
   * Gets the next page using the given function with a basic protocol:
   * Append the page number, page size, success, and error handlers to the provided static arguments.
   * After the error handler, a final handler is provided.
   * The page handler is then called with the final argument list if it exists.
   * The arguments are then removed from the static argument list.
   */
  pager.prototype.nextPage = function () {
    var params = this.settings.loadPageParameters;
    var settings = this.settings;

    params.push(settings.currentPage ++);   // Push page and advance page number
    params.push(settings.pageSize);
    params.push(settings.loadPageSuccess);
    params.push(settings.loadPageError);

    if (settings.loadPageFunction != null) {
        settings.loadPageFunction.apply(settings.loadPageApi, params);
    }

    params.splice(-4, 4);
  };

  /**
   * Sets the current page number
   * page: the new page to set the pager to
   */
  pager.prototype.setPage = function (page) {
    this.settings.currentPage = page;
  };

  /**
   * creates a default loading icon
   * returns: an HMTLElement containing a loading icon
   */
  var createLoadingIcon = function () {
    var icon = document.createElement("span");
    icon.appendChild(document.createTextNode("Loading..."));
    icon.className = "pager-loading-icon";
    return icon;
  };

  /**
   * combines two functions and returns a wrapper that calls both
   * returns: if both functions are defined, a wrapper that calls both with the current context and arguments
   *          if one of the two arguments is undefined, returns the one that is defined
   *          if neither function is defined, returns undefined
   */
  var coupleFunctions = function (a, b) {
    if (a == null) return b;
    if (b == null) return a;
    return function () {
      a.apply(this, arguments);
      b.apply(this, arguments);
    };
  };

  window.Pager = pager;

})( jQuery );
