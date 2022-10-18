;(function (window) {
  function GoDebug() {
    var _godebug = {}

    _godebug.createDebugButton = function () {
      const container = document.createElement("div")
      container.style.zIndex = "9999"
      container.style.background = "var(--bg-color)"
      container.style.bottom = "10px"
      container.style.right = "10px"
      container.style.position = "fixed"

      const button = document.createElement("button")

      button.style.color = "var(--font-color)"
      button.style.padding = ".5rem 1.2rem"
      button.style.fontSize = "1rem"
      button.style.borderRadius = "0.5em"
      button.style.background = "#ffaa31"
      button.style.border = "1px solid #ffaa31"
      button.style.transition = "all .3s"
      button.style.boxShadow = "6px 6px 12px #c5c5c5 -6px -6px 12px #ffffff"
      button.id = "debug-view-button"

      button.innerHTML = "View Data"

      button.addEventListener("click", function () {
        const dialog = document.getElementById("dialog-debug-modal")
        if (typeof dialog.showModal !== "function") {
          console.error("browser does not support dialogs")
        } else {
          dialog.showModal()
        }
      })

      container.appendChild(button)

      return container
    }

    _godebug.createDebugModal = function () {
      const dialog = document.createElement("dialog")
      dialog.id = "dialog-debug-modal"
      const buttonContainer = document.createElement("div")
      buttonContainer.style.display = "flex"
      buttonContainer.style.justifyContent = "center"

      const close = document.createElement("button")
      close.innerHTML = "close"
      close.style.margin = ".5rem"
      close.addEventListener("click", () => {
        dialog.close()
      })

      const render = document.createElement("button")
      render.innerHTML = "render"
      render.style.display = "none"
      render.style.margin = ".5rem"
      render.id = "debug-render"
      render.setAttribute("disabled", true)
      render.addEventListener("click", () => {
        dialog.close()
      })

      buttonContainer.appendChild(close)
      buttonContainer.appendChild(render)

      const editorContainer = document.createElement("div")
      editorContainer.style.width = "800px"
      editorContainer.style.height = "600px"
      editorContainer.style.border = "1px solid grey"
      editorContainer.id = "debug-editor"

      dialog.appendChild(editorContainer)
      dialog.appendChild(buttonContainer)

      return dialog
    }

    // _godebug.createMonacoDependency = function () {
    //     const script =  document.createElement('script')
    //     // script.src = "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.33.0/min/vs/loader.min.js"
    //     // script.setAttribute("integrity", "sha512-O9SYDgWAM3bEzit1z6mkFd+dxKUplO/oB8UwYGAkg2Zy/WzDUQ2mYA/ysk3c0CxiXAN4u8T9JeZ0Ahk2Jj/33Q==")
    //     // script.setAttribute("crossorigin", "anonymous")
    //     // script.setAttribute("referrerpolicy", "no-referer")
    //     script.src = "/monaco-editor/min/vs/loader.js"
    //     return script
    // }

    _godebug.createWasmExec = function () {
      const script = document.createElement("script")
      script.src = "/js/wasm_exec.js"
      return script
    }

    // renderEditorContent renders the json data passed in the debug-editor container
    _godebug.renderEditorContent = function (data) {
      window.require.config({ paths: { vs: "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.33.0/min/vs" } })
      window.require(["vs/editor/editor.main"], function () {
        var editor = monaco.editor.create(document.getElementById("debug-editor"), {
          value: JSON.stringify(data, "  ", 2), //['function x() {', '\tconsole.log("Hello world!");', '}'].join('\n'),
          language: "json",
          folding: true,
        })
        // editor.trigger('fold', 'editor.foldAll')
        editor.getAction("editor.foldLevel2").run()

        //   $('#debug-editor').resize(function(){
        //     editor.layout();
        //   });

        editor.getModel().onDidChangeContent((event) => {
          // console.log('content changed: ', event)
          // console.log('value:', editor.getValue())
        })

        document.getElementById("debug-view-button").addEventListener("click", () => {
          editor.layout()
        })

        // let render = document.getElementById('debug-render')
        // if (render) {
        //     render.addEventListener('click', () => {
        //         let data = editor.getValue()
        //         renderNewTemplate(data)
        //     })
        //     render.removeAttribute('disabled')
        // }
      })
    }

    _godebug.initWasmGO = function () {
      let go = new Go()
      WebAssembly.instantiateStreaming(fetch("/js/main.wasm"), go.importObject).then((result) => {
        go.run(result.instance)
      })
    }

    // async function renderNewTemplate(data) {
    //   console.log("template data", data)

    //   let templates = JSON.parse(data).DebugTemplates
    //   console.log("rendering templates", templates)

    //   const tmplRequests = []
    //   for (let i = 0; i < templates.length; i++) {
    //     tmplRequests.push(fetch("/" + templates[i]))
    //   }

    //   const results = await Promise.all(tmplRequests)

    //   const tmplResults = []
    //   for (let i = 0; i < results.length; i++) {
    //     tmplResults.push(await results[i].text())
    //   }

    //   console.log("results:", tmplRequests)

    //   try {
    //     // data JSON.stringify({ Data: { Title: "friendly world" }})
    //     let tmpl = renderTemplate(tmplResults.join("\n"), data)
    //     console.log("templ rendered: ", tmpl)
    //     var parser = new DOMParser()
    //     var htmlDoc = parser.parseFromString(tmpl, "text/html")
    //     // console.log('SCRIPTS',htmlDoc.scripts)
    //     document.documentElement.replaceWith(htmlDoc.documentElement)
    //     for (let script of Array.from(document.scripts)) {
    //       let ns = document.createElement("script")
    //       if (script.src) {
    //         continue
    //       }
    //       console.log("executing script: ", script)
    //       ns.innerHTML = "(function scope(){" + script.innerHTML + "})()"
    //       document.body.appendChild(ns)
    //     }
    //   } catch (err) {
    //     console.error("err parsing text", err)
    //   }
    // }

    _godebug.initialize = function (data) {
      // document.body.appendChild(this.createWasmExec())

      if (typeof window.require != "function") {
        console.error("the debug library requires the monaco-editor loader library")
        return
      }
      document.body.appendChild(this.createDebugButton())
      document.body.appendChild(this.createDebugModal())
      // console.log('rendering editor content: ', data)
      this.renderEditorContent(data)
      // window.addEventListener('load', () => {
      //     console.log('page loaded')
      //     this.initWasmGO()
      // })
    }

    return _godebug
  }

  window.GoDebug = GoDebug()
})(window)
