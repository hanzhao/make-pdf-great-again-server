'0use strict';

// This script file is used for injecting our code to PDF.js viewer.
(function (global) {
  global.makePdfGreatAgain = {
    textLayerBuilders: [],
    matches: [],
    filename: global.location.search.split('/')[4],
    // Render highlight in the pdf controller
    renderHighlight: function renderHighlight(textLayerBuilder) {
      textLayerBuilder.findController = null
      this.matches[textLayerBuilder.pageIdx] = this.matches[textLayerBuilder.pageIdx] || []
      textLayerBuilder.renderMatches(this.matches[textLayerBuilder.pageIdx])
      this.matches[textLayerBuilder.pageIdx] = []
      this.textLayerBuilders[textLayerBuilder.pageIdx] = textLayerBuilder
    },
    // Rerender all highlights of PDF file.
    rerenderAllHighlight: function rerenderAllHighlight() {
      for (var i = 0; i < this.textLayerBuilders.length; ++i) {
        if (!this.textLayerBuilders[i]) {
          continue;
        }
        this.renderHighlight(this.textLayerBuilders[i])
      }
    },
    // Highlight current selection.
    highlightSelection: function highlightSelection() {
      var selection = global.getSelection()
      var beginTextDiv = selection.anchorNode.parentNode
      var endTextDiv = selection.focusNode.parentNode
      var beginPage = -1, endPage = -1, beginIdx = -1, endIdx = -1
      var beginOffset = selection.anchorOffset, endOffset = selection.focusOffset
      var hs
      // Find begin page and div
      for (var i = 0; i < this.textLayerBuilders.length; ++i) {
        if (!this.textLayerBuilders[i]) {
          continue;
        }
        for (var j = 0; j < this.textLayerBuilders[i].textDivs.length; ++j) {
          if (this.textLayerBuilders[i].textDivs[j] == beginTextDiv) {
            beginPage = i, beginIdx = j
          }
          if (this.textLayerBuilders[i].textDivs[j] == endTextDiv) {
            endPage = i, endIdx = j
          }
        }
      }
      if (beginPage == -1 || endPage == -1) {
        return
      }
      // Swap if selecting from end to begin
      if (beginPage > endPage || (beginPage == endPage && beginIdx > endIdx) ||
        (beginPage == endPage && beginIdx == endIdx && beginOffset > endOffset)) {
        var tmp
        tmp = beginPage; beginPage = endPage; endPage = tmp
        tmp = beginIdx; beginIdx = endIdx; endIdx = tmp
        tmp = beginOffset; beginOffset = endOffset; endOffset = tmp
      }
      this.matches[beginPage] = this.matches[beginPage] || []
      this.matches[endPage] = this.matches[endPage] || []
      // In the same page
      if (beginPage == endPage) {
        hs = [{
          page: beginPage,
          begin: { divIdx: beginIdx, offset: beginOffset },
          end: { divIdx: endIdx, offset: endOffset },
        }]
      } else {
        // In different page, change to two highlight.
        hs = [{
          page: beginPage,
          begin: { divIdx: beginIdx, offset: beginOffset },
          end: {
            divIdx: this.textLayerBuilders[beginPage].textDivs.length - 1,
            offset: 99999999,
          },
        }, {
          page: endPage,
          begin: { divIdx: 0, offset: 0 },
          end: { divIdx: endIdx, offset: endOffset },
        }]
      }
      selection.empty()
      return this.makeHighlight(hs)
    },
    // Use a websocket to watching highlights of the PDF file.
    startWSConnection: function startWSConnection() {
      var ctx = this
      var location = global.location
      var uri = 'ws:'
      if (location.protocol === 'https:') {
        uri = 'wss:'
      }
      uri += '//' + location.host + '/watch/' + ctx.filename
      ctx.ws = new WebSocket(uri)
      ctx.ws.onopen = function () {
        console.log('WS Connected: ' + ctx.filename)
      }
      ctx.ws.onmessage = function (e) {
        var hs = JSON.parse(e.data), page = -1
        for (var i = 0; i < hs.length; ++i) {
          page = hs[i].page
          ctx.matches[page] = ctx.matches[page] || []
          ctx.matches[page].push(hs[i])
        }
        ctx.rerenderAllHighlight()
      }
    },
    // Connnet server to make a new highlight.
    makeHighlight: function makeHighlight(hs) {
      console.log(this.filename)
      for (var i = 0; i < hs.length; ++i) {
        global.fetch('/highlight/' + this.filename, {
          method: 'POST',
          body: JSON.stringify(hs[i]),
          headers: new Headers({
            'Content-Type': 'application/json',
          })
        })
      }
    },
  }
  // Start websocket
  global.makePdfGreatAgain.startWSConnection()
})(window)
