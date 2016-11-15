'use strict';

(function (global) {
  global.makePdfGreatAgain = {
    textLayerBuilders: [],
    matches: [],
    // Render highlight in the pdf controller
    renderHighlight: function renderHighlight(textLayerBuilder) {
      textLayerBuilder.findController = null
      textLayerBuilder.renderMatches(this.matches[textLayerBuilder.pageIdx] || [])
      this.textLayerBuilders[textLayerBuilder.pageIdx] = textLayerBuilder
    },
    rerenderAllHighlight: function rerenderAllHighlight() {
      for (var i = 0; i < this.textLayerBuilders.length; ++i) {
        if (!this.textLayerBuilders[i]) {
          continue;
        }
        this.renderHighlight(this.textLayerBuilders[i])
      }
    },
    highlightSelection: function highlightSelection() {
      var selection = global.getSelection()
      var beginTextDiv = selection.anchorNode.parentNode
      var endTextDiv = selection.focusNode.parentNode
      var beginPage = -1, endPage = -1, beginIdx = -1, endIdx = -1
      var beginOffset = selection.anchorOffset, endOffset = selection.focusOffset
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
      // Swap if selecting from end to begin
      if (beginPage > endPage || (beginPage == endPage && beginIdx > endIdx)) {
        var tmp
        tmp = beginPage; beginPage = endPage; endPage = tmp
        tmp = beginIdx; beginIdx = endIdx; endIdx = tmp
        tmp = beginOffset; beginOffset = endOffset; endOffset = tmp
      }
      this.matches[beginPage] = this.matches[beginPage] || []
      this.matches[endPage] = this.matches[endPage] || []
      // In the same page
      if (beginPage == endPage) {
        this.matches[beginPage].push({
          begin: { divIdx: beginIdx, offset: beginOffset },
          end: { divIdx: endIdx, offset: endOffset },
        })
      } else {
        this.matches[beginPage].push({
          begin: { divIdx: beginIdx, offset: beginOffset },
          end: {
            divIdx: this.textLayerBuilders[beginPage].textDivs.length - 1,
            offset: Infinity,
          },
        })
        this.matches[endPage].push({
          begin: { divIdx: 0, offset: 0 },
          end: { divIdx: endIdx, offset: endOffset },
        })
      }
      selection.empty()
      // render
      return this.rerenderAllHighlight()
    }
  }
})(window)
