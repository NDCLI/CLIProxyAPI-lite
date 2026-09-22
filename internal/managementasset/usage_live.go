package managementasset

import "bytes"

// AddUsageLivePolling installs an authenticated SSE client for live usage updates.
func AddUsageLivePolling(html []byte) ([]byte, bool) {
	if !bytes.Contains(html, []byte("cliproxy-usage-history")) || bytes.Contains(html, []byte("cliproxy-usage-live-poll")) {
		return html, false
	}
	const extension = `<script id="cliproxy-usage-live-poll">
(()=>{let controller=null;async function stream(){if(controller||!location.hash.includes("#/usage-history"))return;const auth=window.__cliproxyUsageAuth;if(!auth)return;controller=new AbortController();try{const r=await fetch("/v0/management/usage-history/stream",{headers:{Authorization:auth},signal:controller.signal});if(!r.ok||!r.body)throw new Error();const reader=r.body.getReader(),decoder=new TextDecoder();let buffer="";for(;;){const part=await reader.read();if(part.done)break;buffer+=decoder.decode(part.value,{stream:true});const events=buffer.split("\n\n");buffer=events.pop()||"";for(const event of events){const line=event.split("\n").find(x=>x.startsWith("data: "));if(line&&window.__cliproxyUsageRefresh){window.__cliproxyUsageRefresh(JSON.parse(line.slice(6)))}}}}catch(_){}finally{controller=null;if(location.hash.includes("#/usage-history"))setTimeout(stream,1500)}}function tick(){if(!location.hash.includes("#/usage-history")){controller?.abort();controller=null;return}stream()}window.addEventListener("hashchange",tick);setInterval(tick,1000);tick()})();
</script>`
	return bytes.Replace(html, []byte("</body>"), append([]byte(extension), []byte("</body>")...), 1), true
}
