package managementasset

import "bytes"

// AddEndpointCard adds the local OpenAI-compatible endpoint as a normal card
// in the management page's configuration content.
func AddEndpointCard(html []byte) ([]byte, bool) {
	if !bytes.Contains(html, []byte("</body>")) || bytes.Contains(html, []byte("cliproxy-endpoint-card")) {
		return html, false
	}
	const extension = `<style>
#cliproxy-endpoint-card{box-sizing:border-box;width:100%;margin:16px 0;padding:16px;background:#211f1c;color:#f7f4ef;border:1px solid #4a4640;border-radius:12px;font:13px system-ui,sans-serif}
#cliproxy-endpoint-card strong{display:block;margin-bottom:8px;font-size:15px}
#cliproxy-endpoint-card .cliproxy-endpoint-row{display:flex;gap:8px;align-items:center}
#cliproxy-endpoint-card code{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:#b8f1d8}
#cliproxy-endpoint-card button{border:1px solid #6d675f;border-radius:7px;padding:6px 9px;background:#35312d;color:#fff;cursor:pointer}
</style>
<script>
(()=>{function onConfigTab(){return /#\/config(?:[/?#]|$)/.test(location.hash)&&/Bảng cấu hình|Configuration/i.test(document.body?.innerText||"")}function mount(){const old=document.getElementById("cliproxy-endpoint-card");if(!onConfigTab()){old?.remove();return}if(old)return;const main=document.querySelector("main")||document.body;const card=document.createElement("section");card.id="cliproxy-endpoint-card";const endpoint=location.origin+"/v1";card.innerHTML='<strong>API Endpoint</strong><div class="cliproxy-endpoint-row"><code></code><button type="button">Sao chép</button></div>';card.querySelector("code").textContent=endpoint;card.querySelector("button").addEventListener("click",async()=>{try{await navigator.clipboard.writeText(endpoint);card.querySelector("button").textContent="Đã sao chép";setTimeout(()=>card.querySelector("button").textContent="Sao chép",1500)}catch(_){}});main.appendChild(card)}window.addEventListener("hashchange",mount);setInterval(mount,500);if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",mount);else mount()})();
</script>`
	return bytes.Replace(html, []byte("</body>"), append([]byte(extension), []byte("</body>")...), 1), true
}
