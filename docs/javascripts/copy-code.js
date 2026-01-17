;(function () {
  function addCopyButtons() {
    const blocks = document.querySelectorAll('pre > code')
    blocks.forEach((codeBlock) => {
      const pre = codeBlock.parentNode
      if (pre.classList.contains('has-copy-button')) return

      const button = document.createElement('button')
      button.className = 'copy-code-button'
      button.type = 'button'
      button.innerText = 'Copy'

      button.addEventListener('click', async () => {
        const text = codeBlock.innerText
        try {
          await navigator.clipboard.writeText(text)
          button.innerText = 'Copied'
          setTimeout(() => (button.innerText = 'Copy'), 1500)
        } catch (err) {
          console.error('Copy failed', err)
          button.innerText = 'Error'
          setTimeout(() => (button.innerText = 'Copy'), 1500)
        }
      })

      const wrapper = document.createElement('div')
      wrapper.className = 'code-block-wrapper'
      pre.classList.add('wrap-lines')
      pre.parentNode.insertBefore(wrapper, pre)
      wrapper.appendChild(pre)
      const actions = document.createElement('div')
      actions.className = 'code-actions'
      actions.appendChild(button)
      wrapper.appendChild(actions)
      pre.classList.add('has-copy-button')
    })
  }

  document.addEventListener('DOMContentLoaded', addCopyButtons)
  document.addEventListener('DOMContentLoaded', () => {
    const observer = new MutationObserver(addCopyButtons)
    observer.observe(document.body, { childList: true, subtree: true })
  })
})()
