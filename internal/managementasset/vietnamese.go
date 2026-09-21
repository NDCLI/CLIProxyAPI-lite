package managementasset

import (
	"bytes"
	_ "embed"
)

//go:embed vi-dom.json
var vietnameseDOM string

const vietnameseBootstrap = `<script>
(()=>{const translations=__VI_TRANSLATIONS__;const storageKey="cliproxy-dashboard-language";let language=localStorage.getItem(storageKey)==="vi"?"vi":"en";let original=new WeakMap;let observer;function textNodes(root){const walker=document.createTreeWalker(root,NodeFilter.SHOW_TEXT);const nodes=[];while(walker.nextNode())nodes.push(walker.currentNode);return nodes}function translateText(value){if(language!=="vi")return value;const trimmed=value.trim(),replacement=translations[trimmed];if(!replacement)return value;const left=value.slice(0,value.indexOf(trimmed)),right=value.slice(value.indexOf(trimmed)+trimmed.length);return left+replacement+right}function apply(root=document.body){if(!root)return;textNodes(root).forEach(node=>{if(!original.has(node))original.set(node,node.nodeValue);node.nodeValue=translateText(original.get(node))});document.documentElement.lang=language==="vi"?"vi":"en"}function closeMenu(){document.getElementById("cliproxy-language-menu")?.remove()}function toggleMenu(button){const existing=document.getElementById("cliproxy-language-menu");if(existing){closeMenu();return}const box=button.getBoundingClientRect(),menu=document.createElement("div");menu.id="cliproxy-language-menu";menu.setAttribute("role","menu");menu.style.cssText="position:fixed;z-index:100000;left:"+box.left+"px;top:"+(box.bottom+6)+"px;min-width:156px;padding:6px;background:#242424;color:#fff;border:1px solid #555;border-radius:8px;box-shadow:0 10px 30px rgba(0,0,0,.45);font:13px system-ui";for(const [id,label] of [["en","English"],["vi","Tiếng Việt"]]){const item=document.createElement("button");item.type="button";item.textContent=(language===id?"✓ ":"")+label;item.style.cssText="display:block;width:100%;padding:8px 10px;background:transparent;border:0;border-radius:5px;color:inherit;text-align:left;cursor:pointer";item.onmouseenter=()=>item.style.background="#383838";item.onmouseleave=()=>item.style.background="transparent";item.onclick=()=>{language=id;localStorage.setItem(storageKey,language);apply();closeMenu()};menu.appendChild(item)}document.body.appendChild(menu)}function setup(){document.addEventListener("click",event=>{const button=event.target.closest('button[aria-label="Language"],button[title="Language"]');if(button){event.preventDefault();event.stopImmediatePropagation();toggleMenu(button);return}if(!event.target.closest("#cliproxy-language-menu"))closeMenu()},true);apply();observer?.disconnect();observer=new MutationObserver(mutations=>{for(const mutation of mutations){mutation.addedNodes.forEach(node=>{if(node.nodeType===Node.TEXT_NODE||node.nodeType===Node.ELEMENT_NODE)apply(node.nodeType===Node.TEXT_NODE?node.parentElement:node)})}});observer.observe(document.body,{childList:true,subtree:true})}if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",setup);else setup()})();
</script>`

// AddVietnameseLocale leaves the upstream English dashboard untouched and adds
// a two-option, local-only language picker. This stays compatible with updated,
// minified upstream management assets because it does not patch their i18n code.
func AddVietnameseLocale(html []byte) ([]byte, bool) {
	if !bytes.Contains(html, []byte("<body")) || !bytes.Contains(html, []byte("</body>")) {
		return html, false
	}
	bootstrap := bytes.Replace([]byte(vietnameseBootstrap), []byte("__VI_TRANSLATIONS__"), []byte(vietnameseDOM), 1)
	return bytes.Replace(html, []byte("</body>"), append(bootstrap, []byte("</body>")...), 1), true
}
