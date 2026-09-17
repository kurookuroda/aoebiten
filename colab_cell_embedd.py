%%writefile /content/merge_wasm8.py
"""

cd aoebiten
pwd
go get github.com/hajimehoshi/ebiten/v2@latest
# go mod init はすでに go.mod が存在するため不要。削除。
go mod tidy
#go build -tags embed -o aozora-reader
#./aozora-reader
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -tags embed -o aozora-reader.wasm

cp "/content/aoebiten/aozora-reader.wasm" /content/
cp "/content/aoebiten/index.html" /content/
python3 merge_wasm8.py wasm_exec.js aozora-reader.wasm


"""
import base64
import sys
import os

def merge_to_single_html(wasm_exec_path="wasm_exec.js", wasm_path="aozora-reader.wasm", output_path="aozora-reader-single.html"):
    if not os.path.exists(wasm_exec_path):
        print(f"Error: {wasm_exec_path} not found.")
        sys.exit(1)
    if not os.path.exists(wasm_path):
        print(f"Error: {wasm_path} not found.")
        sys.exit(1)

    with open(wasm_exec_path, "r", encoding="utf-8") as f:
        wasm_exec_js = f.read()
    wasm_exec_js = wasm_exec_js.replace("</script>", "<\\/script>")

    with open(wasm_path, "rb") as f:
        wasm_bytes = f.read()
    wasm_b64 = base64.b64encode(wasm_bytes).decode("ascii")

    html_content = f"""<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>8 ビットポートピア</title>
<style>
*, *::before, *::after {{
    box-sizing: border-box;
}}
html, body {{
    margin: 0;
    padding: 0;
    background-color: #111;
    color: #ccc;
    font-family: sans-serif;
    overflow-x: hidden;
    height: 100%;
}}
body {{
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    min-height: 100%;
    padding: 0.5rem;
}}
h2 {{
    font-size: 1.1rem;
    margin: 0.3rem 0;
    flex-shrink: 0;
}}
#game-container {{
    position: relative;
    width: 100%;
    max-width: 100%;
    max-height: 70vh;
    aspect-ratio: 1 / 1;
    border: 2px solid #555;
    image-rendering: pixelated;
    overflow: hidden;
    flex-shrink: 0;
    background-color: #000;
}}
#game-container canvas {{
    display: block;
    position: absolute !important;
    left: 0 !important;
    top: 0 !important;
    width: 100% !important;
    height: 100% !important;
    image-rendering: pixelated;
}}
#loading {{
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 1rem;
    pointer-events: none;
    background: rgba(0,0,0,0.9);
    padding: 1rem;
    z-index: 10;
}}
.controls {{
    margin-top: 0.5rem;
    font-size: 0.75rem;
    line-height: 1.5;
    max-width: 100%;
    width: 100%;
    flex-shrink: 0;
}}
.controls h2 {{
    font-size: 0.9rem;
    margin: 0.3rem 0;
}}
.controls table {{
    border-collapse: collapse;
    width: 100%;
}}
.controls th, .controls td {{
    border: 1px solid #444;
    padding: 2px 4px;
    text-align: left;
}}
.controls th {{
    background: #222;
}}
.controls p {{
    margin: 0.3rem 0;
    font-size: 0.7rem;
}}
</style>
</head>
<body>
<h2>8 ビットポートピア</h2>
<div id="game-container">
    <canvas id="placeholder" style="display:none;"></canvas>
    <div id="loading">Now Loading...</div>
</div>

<div class="controls">
    <h2>操作方法</h2>
    <table>
        <tr><th>機能</th><th>キー/操作</th></tr>
        <tr><td>進む / 決定</td><td>タップ / Enter / ↓</td></tr>
        <tr><td>スキップ</td><td>タップ / Enter / ↓</td></tr>
        <tr><td>高速送り</td><td>↓ + Z</td></tr>
        <tr><td>超高速送り</td><td>↓ + X</td></tr>
        <tr><td>前のページ</td><td>↑</td></tr>
        <tr><td>フォント切替</td><td>F</td></tr>
        <tr><td>リセット</td><td>R</td></tr>
    </table>
    <p>※ タップで進めます。キーボード接続でより快適に操作可能です。</p>
</div>

<script>
{wasm_exec_js}
</script>

<script>
// ===== 音声対応 =====
window.__game_audio_context__ = null;

function initAudioContext() {{
    console.log("initAudioContext called");
    if (!window.__game_audio_context__) {{
        const AudioContextClass = window.AudioContext || window.webkitAudioContext;
        if (AudioContextClass) {{
            window.__game_audio_context__ = new AudioContextClass();
            console.log("AudioContext created, state:", window.__game_audio_context__.state);
        }} else {{
            console.error("AudioContext not supported");
        }}
    }}
    if (window.__game_audio_context__) {{
        if (window.__game_audio_context__.state === "suspended") {{
            window.__game_audio_context__.resume().then(() => {{
                console.log("AudioContext resumed, state:", window.__game_audio_context__.state);
            }}).catch(err => console.error("resume failed:", err));
        }} else {{
            console.log("AudioContext already state:", window.__game_audio_context__.state);
        }}
    }}
}}

// キャプチャフェーズで複数イベントを監視
["click", "touchstart", "touchend", "keydown"].forEach(evt => {{
    document.addEventListener(evt, initAudioContext, {{ capture: true }});
}});

// Goから呼び出される音声関数
window.playBeep = function(freq, durationSec) {{
    console.log("playBeep called:", freq, durationSec);
    if (!window.__game_audio_context__) {{
        console.log("playBeep: AudioContext not ready, trying to create...");
        initAudioContext();
        if (!window.__game_audio_context__) {{
            console.error("playBeep: failed to create AudioContext");
            return;
        }}
    }}

    const ctx = window.__game_audio_context__;
    console.log("playBeep: AudioContext state:", ctx.state);

    // suspended なら resume を試みる
    if (ctx.state === "suspended") {{
        ctx.resume();
        console.log("playBeep: called resume()");
    }}

    const now = ctx.currentTime;
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = "square";
    osc.frequency.value = freq;

    gain.gain.setValueAtTime(0.1, now);
    gain.gain.exponentialRampToValueAtTime(0.01, now + durationSec);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start(now);
    osc.stop(now + durationSec);

    console.log("playBeep: started oscillator at", freq, "Hz");
}};
// =====================

const wasmBase64 = "{wasm_b64}";

function base64ToUint8Array(base64) {{
    const binary_string = window.atob(base64);
    const len = binary_string.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {{
        bytes[i] = binary_string.charCodeAt(i);
    }}
    return bytes;
}}

const go = new Go();
const wasmBytes = base64ToUint8Array(wasmBase64);
const blob = new Blob([wasmBytes], {{ type: "application/wasm" }});
const url = URL.createObjectURL(blob);

const container = document.getElementById("game-container");
const loading = document.getElementById("loading");

const observer = new MutationObserver((mutations) => {{
    for (const mutation of mutations) {{
        for (const node of mutation.addedNodes) {{
            if (node.tagName === "CANVAS") {{
                node.style.position = "absolute";
                node.style.left = "0";
                node.style.top = "0";
                node.style.width = "100%";
                node.style.height = "100%";
                container.appendChild(node);
                loading.style.display = "none";
                observer.disconnect();
            }}
        }}
    }}
}});
observer.observe(document.body, {{ childList: true, subtree: true }});

WebAssembly.instantiateStreaming(fetch(url), go.importObject)
    .then((result) => {{
        go.run(result.instance);
    }})
    .catch((err) => {{
        loading.textContent = "Error: " + err;
        console.error(err);
        observer.disconnect();
    }});
</script>
</body>
</html>"""

    with open(output_path, "w", encoding="utf-8") as f:
        f.write(html_content)

    original_size = os.path.getsize(wasm_path)
    output_size = os.path.getsize(output_path)
    print(f"Generated: {output_path}")
    print(f"Original WASM: {original_size:,} bytes ({original_size/1024/1024:.2f} MB)")
    print(f"Single HTML:   {output_size:,} bytes ({output_size/1024/1024:.2f} MB)")
    print(f"Size ratio:    {output_size/original_size:.2f}x")

if __name__ == "__main__":
    if len(sys.argv) >= 3:
        merge_to_single_html(sys.argv[1], sys.argv[2])
    else:
        merge_to_single_html()
