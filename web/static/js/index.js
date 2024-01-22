import hljs from './highlight.js';
import go from './go.min.js';
import {marked} from "./marked.esm.js";

let messages = []
let messagesDiv

window.onload = () => {
    initHLJS()
    initPanicDiv()
    initMessages()
    initInput()
}

function initHLJS() {
    hljs.registerLanguage('go', go)
    hljs.configure({
        'ignoreUnescapedHTML': true
    })
}

function initPanicDiv() {
    axios.get("/api/panic").then(res => {
        const data = res.data.data
        const panicTitleElement = document.getElementById('panic-title')
        const panicTracebackElement = document.getElementById('panic-traceback')
        panicTitleElement.innerText = data['panic']

        const hoverElement = createElement('div', 'panic-traceback-hover')
        const hoverElementPre = createElement('pre', 'panic-traceback-hover-pre')
        const hoverElementCode = createElement('code', 'panic-traceback-hover-code')
        const hoverElementFooter = createElement('div', 'panic-traceback-hover-footer')
        let appElement = document.getElementById("app")
        hoverElementPre.append(hoverElementCode)
        hoverElement.append(hoverElementPre)
        hoverElement.append(hoverElementFooter)
        hoverElement.style.display = 'none'
        hoverElementCode.classList.add('language-go')
        appElement.append(hoverElement)
        appElement.onclick = (event) => {
            let element = event.target
            for (let element = event.target; element.localName !== 'body'; element = element.parentElement) {
                if (element.className === 'panic-traceback' || element.className === 'panic-traceback-hover') {
                    return
                }
            }
            hoverElement.style.display = 'none'
        }

        for (let i in data['functions']) {
            const func = data['functions'][i]
            if (!func['source'].length) func['source'] = 'This function does not support preview yet.'
            const item = createElement('div', 'panic-traceback-item')
            const itemHeader = createElement('div', 'panic-traceback-item-header')
            const itemOperate = createElement('div', 'panic-traceback-item-operate')
            const itemType = createElement('div', 'panic-traceback-item-type')
            const itemFunc = createElement('div', 'panic-traceback-item-func')
            const itemFooter = createElement('div', 'panic-traceback-item-footer')
            const itemSource = createElement('div', 'panic-traceback-item-source')
            const itemSourcePre = createElement('pre', 'panic-traceback-item-source-pre')
            const itemSourceCode = createElement('code', 'panic-traceback-item-source-code')
            const itemFile = createElement('a', 'panic-traceback-item-source-file')
            item.append(itemHeader)
            itemHeader.append(itemOperate)
            itemHeader.append(itemType)
            itemHeader.append(itemFunc)
            item.append(itemFooter)
            itemFooter.append(itemSource)
            itemSource.append(itemSourcePre)
            itemSourcePre.append(itemSourceCode)
            itemSource.append(itemFile)
            panicTracebackElement.append(item)

            itemType.innerText = func['type'][0]
            itemType.classList.add(`panic-traceback-item-type-${func['type']}`)
            let funcDefine = `<span class="panic-traceback-item-func-name">${func['name']}</span>`
            funcDefine += '('
            if (func['params']) {
                let params = []
                for (let param of func['params']) {
                    let paramHTML = param['name']
                    if (param['type'].length > 0)
                        paramHTML += ` <span class="panic-traceback-item-func-field-type">${param['type']}</span>`
                    params.push(paramHTML)
                }
                funcDefine += params.join(', ')
            }
            funcDefine += ')'
            if (func['results']) {
                let results = []
                for (let result of func['results'])
                    if (result['type'].length > 0)
                        results.push(`<span class="panic-traceback-item-func-field-type">${result['type']}</span>`)
                funcDefine += ' '
                if (results.length > 1) funcDefine += '('
                funcDefine += results.join(', ')
                if (results.length > 1) funcDefine += ')'
            }
            itemFunc.innerHTML = funcDefine
            itemFile.href = '/files' + func['file']
            itemFile.innerText = func['file']
            itemSourceCode.innerHTML = func['source']
            itemSourceCode.classList.add('language-go')
            highlightElement(itemSourceCode, false);

            const openClass = 'panic-traceback-item-open'
            const itemFooterHeight = itemFooter.scrollHeight + 'px'
            item.setAttribute('status', i !== '0')
            const onClick = function () {
                document.getSelection().removeAllRanges();
                const status = JSON.parse(item.getAttribute('status'))
                item.setAttribute('status', !status)
                if (!status) {
                    item.classList.add(openClass)
                    itemFooter.style.height = itemFooterHeight
                } else {
                    item.classList.remove(openClass)
                    itemFooter.style.height = 0
                }
            }
            onClick()
            itemOperate.onclick = onClick
            itemHeader.ondblclick = onClick

            let hoverTimeOut = null;
            itemHeader.onmousemove = (event) => {
                if (hoverTimeOut) clearTimeout(hoverTimeOut)
                hoverTimeOut = setTimeout(function () {
                    const status = JSON.parse(item.getAttribute('status'))
                    if (status) return
                    hoverElementCode.innerHTML = func['source']
                    highlightElement(hoverElementCode, false, false)
                    hoverElementFooter.innerHTML = `${getBase(func['name'])} <i>(${getBase(func['file'])}:${func['line']})</i>`
                    hoverElement.style.left = event.pageX + 'px'
                    hoverElement.style.top = event.pageY + 'px'
                    hoverElement.style.display = 'unset'
                }, 750);
            }
            itemHeader.onmouseout = () => {
                if (hoverTimeOut) clearTimeout(hoverTimeOut);
            }
        }
    })
}

function initInput() {
    const inputElement = document.getElementById('input')
    const inputInnerElement = document.getElementById('input-inner')
    const inputSubmitElement = document.getElementById('input-submit')

    inputElement.onclick = () => {
        inputInnerElement.focus()
    }

    inputInnerElement.onfocus = () => {
        inputElement.classList.add('input-focus')
    }
    inputInnerElement.onblur = () => {
        inputElement.classList.remove('input-focus')
    }

    inputInnerElement.oninput = () => {
        inputInnerElement.style.height = 'auto'
        inputInnerElement.style.height = inputInnerElement.scrollHeight + 'px'
    }

    inputElement.onsubmit = (e) => {
        e.preventDefault()
        sendMsg(inputInnerElement.value)
        inputInnerElement.value = ''
    }
}

function createElement(tagName, ...classNames) {
    const element = document.createElement(tagName)
    element.classList.add(...classNames)
    return element
}

function initMessages() {
    messagesDiv = document.getElementById('messages')
    sendMsg()
}

function newMessageElement(role) {
    const messageElement = document.createElement('div')
    messageElement.classList.add('message')
    messageElement.classList.add(`message-${role}`)
    messagesDiv.append(messageElement)
    return messageElement
}

function sendMsg(message) {
    if (message) {
        const messageElement = newMessageElement('user')
        messageElement.innerHTML = marked.parse(message)
        messages.push({role: 'user', content: message})
        messagesDiv.scrollTop = messagesDiv.scrollHeight - messagesDiv.offsetHeight
    }
    fetch(
        '/api/chat',
        {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({messages})
        }
    ).then((response) => {
        const reader = response.body.getReader()
        const decoder = new TextDecoder('utf-8')
        const messageElement = newMessageElement('assistant')
        let assistantMessage = ''
        let preHeight = messagesDiv.scrollHeight

        function processStreamResult(result) {
            const chunk = decoder.decode(result.value, {stream: !result.done})

            // 解析chunk
            chunk.split('event:').forEach((group) => {
                group = group.slice(0, -2)
                const msgs = group.split('\ndata:')

                if (msgs.length < 2) {
                    return
                }

                if (msgs[0] === 'error') {
                    // TODO: onerror
                    return
                }
                if (msgs[0] === 'done') {
                    messages.push({role: 'assistant', content: assistantMessage})
                    return
                }

                assistantMessage += msgs[1]
                for (let i = 2; i < msgs.length; i++) {
                    if (msgs[i] === '') assistantMessage += '\n'
                }
                messageElement.innerHTML = marked.parse(assistantMessage)
                // hljs.highlightAll()
                for (let children of messageElement.children) {
                    if (children.localName !== 'pre') continue
                    const codeElement = children.children[0]
                    if (codeElement.localName !== 'code') continue
                    highlightElement(codeElement)
                }
                if (messagesDiv.scrollHeight !== preHeight && messagesDiv.scrollTop >= messagesDiv.scrollHeight - messagesDiv.offsetHeight - 120) {
                    messagesDiv.scrollTop = messagesDiv.scrollHeight - messagesDiv.offsetHeight
                }
                preHeight = messagesDiv.scrollHeight
            })

            if (!result.done) {
                reader.read().then(processStreamResult)
            } else {
                messages.push({role: 'assistant', content: assistantMessage})
            }
        }

        reader.read().then(processStreamResult)
    })
}

function getBase(str) {
    let lis = str.split('/');
    if (lis.length === 0) {
        return ''
    }
    return lis[lis.length - 1]
}

function highlightElement(element, showLang = true, showNumber = true) {
    hljs.highlightElement(element)
    let html = ''
    if (showLang) {
        for (let cls of element.classList) {
            element.classList.add('hljs-added-lang')
            if (cls.startsWith('language-')) {
                const lang = cls.replace('language-', '')
                html += `<div class="lang">${lang}</div>`
            }
        }
    }
    if (showNumber) {
        html += '<ol>'
        const lines = element.innerHTML.split('\n')
        for (let i in lines) {
            const line = lines[i]
            if (Number(i) === lines.length - 1 && line.length === 0) break
            html += `<li><span class="code">${line}</span></li>`
        }
        html += '</ol>'
        element.innerHTML = html
    }
}