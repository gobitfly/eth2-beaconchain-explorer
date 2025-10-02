;(function () {
  // Tree table: use row data attributes to build 2-level hierarchy (entity -> sub-entity)
  function buildTree(table) {
    if (!table || !table.tBodies || !table.tBodies[0]) return
    const rows = Array.from(table.tBodies[0].rows)
    if (rows.length === 0) return

    const entityRows = new Map() // entity -> row (level 0)
    const childrenMap = new Map() // entity -> child rows (level 1)

    for (const row of rows) {
      const entity = (row.dataset.entity || "").trim()
      const sub = (row.dataset.subEntity || "").trim()
      if (!entity) continue
      if (!sub) {
        entityRows.set(entity, row)
      } else {
        if (!childrenMap.has(entity)) childrenMap.set(entity, [])
        childrenMap.get(entity).push(row)
      }
    }

    // Hide all child rows initially
    childrenMap.forEach((children) => {
      for (const child of children) child.style.display = "none"
    })

    // Add caret + indent to entity rows that have children
    entityRows.forEach((row, entity) => {
      const hasChildren = childrenMap.has(entity)
      if (!hasChildren) return

      row.classList.add("has-children")
      row.dataset.expanded = "false"

      const labelCell = row.cells[0]
      if (!labelCell) return

      // Build caret and indent, but keep existing content (anchor, etc.)
      const caret = document.createElement("span")
      caret.className = "tt-caret"
      caret.title = "Expand"

      const wrapper = document.createElement("span")
      wrapper.className = "tt-label"
      wrapper.style.marginLeft = "0px"

      // Insert caret before the existing content
      wrapper.appendChild(caret)
      wrapper.appendChild(document.createTextNode(" "))

      // Move existing children of labelCell into wrapper
      while (labelCell.firstChild) {
        wrapper.appendChild(labelCell.firstChild)
      }
      labelCell.appendChild(wrapper)

      const toggle = function () {
        const expanded = row.dataset.expanded === "true"
        const children = childrenMap.get(entity) || []
        if (expanded) {
          for (const child of children) child.style.display = "none"
          row.dataset.expanded = "false"
          caret.classList.remove("expanded")
          caret.title = "Expand"
        } else {
          for (const child of children) child.style.display = "table-row"
          row.dataset.expanded = "true"
          caret.classList.add("expanded")
          caret.title = "Collapse"
        }
      }

      caret.addEventListener("click", function (e) {
        e.stopPropagation()
        toggle()
      })
      labelCell.addEventListener("click", function (e) {
        if (!(e.target && e.target.closest("a"))) {
          toggle()
        }
      })
    })
  }

  document.addEventListener("DOMContentLoaded", function () {
    const table = document.getElementById("staking-pool-table")
    buildTree(table)
    if (window.BeaconTreemap) {
      window.BeaconTreemap.renderFromScript("entities-treemap", "treemap-data")
    }
  })
})()
