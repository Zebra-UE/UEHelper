// filepath: d:\GitHub\UEHelper\Client\src\crash.ts
import axios from 'axios';
import { Terminal } from 'xterm';
import 'xterm/css/xterm.css';

// 初始化 xterm
const term = new Terminal({
  cols: 120,
  rows: 30,
  convertEol: true,
  rendererType: 'canvas',
  fontFamily: 'monospace',
  fontSize: 12,
});

const termEl = document.getElementById('term') as HTMLElement | null;
if (!termEl) {
  throw new Error('missing #term element in DOM');
}
term.open(termEl);

const itemsEl = document.getElementById('items') as HTMLElement | null;
const statusEl = document.getElementById('status') as HTMLElement | null;
const clearBtn = document.getElementById('clearBtn') as HTMLButtonElement | null;

function setStatus(text: string) {
  if (statusEl) statusEl.textContent = text;
}

if (itemsEl) {
  itemsEl.addEventListener('click', (ev) => {
    const target = ev.target as HTMLElement;
    const li = target.closest('.list-item') as HTMLElement | null;
    if (!li) return;

    // 高亮
    itemsEl.querySelectorAll('.list-item').forEach(n => n.classList.remove('active'));
    li.classList.add('active');

    const id = li.getAttribute('data-id');
    if (!id) return;

    setStatus('请求中: ' + id);
    term.clear();
    term.writeln('请求参数: ' + id);

    axios.post('/api/crash', { id })
      .then(resp => {
        setStatus('已返回');
        try {
          const json = resp.data;
          const pretty = JSON.stringify(json, null, 2);
          pretty.split('\n').forEach(line => term.writeln(line));
        } catch (e: any) {
          term.writeln('解析返回数据失败: ' + (e?.message || String(e)));
          term.writeln(String(resp.data));
        }
      })
      .catch(err => {
        setStatus('请求失败');
        if (err.response && err.response.data) {
          term.writeln('错误响应:');
          term.writeln(JSON.stringify(err.response.data, null, 2));
        } else {
          term.writeln('网络或服务器错误: ' + (err.message || String(err)));
        }
      });
  });
}

if (clearBtn) {
  clearBtn.addEventListener('click', () => {
    term.clear();
    setStatus('');
    if (itemsEl) itemsEl.querySelectorAll('.list-item').forEach(n => n.classList.remove('active'));
  });
}