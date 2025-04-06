import tippy from "tippy.js"
import "tippy.js/dist/tippy.css"

export default {
  mounted: function (el, binding, vnode) {
    var text = binding.value.text || "tooltip text"

    tippy(el, {
      content: text,
      allowHTML: false,
    })
  },
}
