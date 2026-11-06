const {test,expect}=require('@playwright/test');
const fs=require('node:fs');
const vm=require('node:vm');

test('context menu reuses the content script before trying injection',async()=>{
  let clicked;
  let injected=0;
  const messages=[];
  const chrome={runtime:{onInstalled:{addListener(){}},onMessage:{addListener(){}}},
    contextMenus:{onClicked:{addListener(fn){clicked=fn}}},
    tabs:{sendMessage:async(id,message)=>{messages.push(message)}},
    scripting:{insertCSS:async()=>{injected++},executeScript:async()=>{injected++}}};
  const source=fs.readFileSync(require.resolve('../background.src.js'),'utf8').replace(/^import .*;\n/,'');
  vm.runInNewContext(source,{chrome,console,URL,SnipoClient:class{}});
  await clicked({menuItemId:'save-to-snipo',selectionText:'code'},{id:1,title:'Example'});
  expect(injected).toBe(0);
  expect(messages).toHaveLength(1);
  expect(messages[0].action).toBe('contextMenuSave');
});

test('a browser with existing scripting access can initialize a missing content script',async()=>{
  let clicked;
  let ready=false;
  let injected=0;
  const messages=[];
  const chrome={runtime:{onInstalled:{addListener(){}},onMessage:{addListener(){}}},
    contextMenus:{onClicked:{addListener(fn){clicked=fn}}},
    tabs:{sendMessage:async(id,message)=>{if(!ready)throw new Error('No receiver');messages.push(message)}},
    scripting:{insertCSS:async()=>{injected++},executeScript:async()=>{injected++;ready=true}}};
  const source=fs.readFileSync(require.resolve('../background.src.js'),'utf8').replace(/^import .*;\n/,'');
  vm.runInNewContext(source,{chrome,console,URL,SnipoClient:class{}});
  await clicked({menuItemId:'save-to-snipo',selectionText:'code'},{id:1,title:'Example'});
  expect(injected).toBe(2);
  expect(messages).toHaveLength(1);
  expect(messages[0].action).toBe('contextMenuSave');
});

test('a browser without scripting access safely quick-saves when no page receiver exists',async()=>{
  let clicked;
  const saved=[];
  const chrome={runtime:{onInstalled:{addListener(){}},onMessage:{addListener(){}}},
    contextMenus:{onClicked:{addListener(fn){clicked=fn}}},
    storage:{sync:{get:async()=>({instanceUrl:'https://snipo.test',apiKey:'test'})}},
    tabs:{sendMessage:async(id,message)=>{if(message.action==='contextMenuSave')throw new Error('No receiver')}}};
  class Client {async createSnippet(payload){saved.push(payload);return {id:'saved'}}}
  const source=fs.readFileSync(require.resolve('../background.src.js'),'utf8').replace(/^import .*;\n/,'');
  vm.runInNewContext(source,{chrome,console,URL,SnipoClient:Client});
  await clicked({menuItemId:'save-to-snipo',selectionText:'    code\n'},{id:1,title:'Example'});
  expect(saved).toHaveLength(1);
  expect(saved[0].files[0].content).toBe('    code\n');
});

test('interactive context-menu saves use an isolated tab when popup windows are unavailable', async () => {
  let clicked;
  const records = {};
  const opened = [];
  const chrome = {
    runtime: { onInstalled: { addListener() {} }, onMessage: { addListener() {} }, getURL: path => 'chrome-extension://fixture/' + path },
    contextMenus: { onClicked: { addListener(listener) { clicked = listener; } } },
    storage: {
      sync: { get: async () => ({ interactiveMode: true }) },
      session: {
        get: async () => ({ ...records }),
        set: async values => Object.assign(records, values),
        remove: async key => { delete records[key]; }
      }
    },
    tabs: {
      sendMessage: async () => { throw new Error('No content script'); },
      create: async options => { opened.push(options); return { id: 2 }; }
    }
  };
  const source = fs.readFileSync(require.resolve('../background.src.js'), 'utf8').replace(/^import .*;\n/, '');
  vm.runInNewContext(source, { chrome, console, URL, crypto: require('node:crypto').webcrypto, SnipoClient: class {} });
  await clicked({ menuItemId: 'save-to-snipo', selectionText: 'private draft' }, { id: 1, title: 'Example', url: 'https://code.test/example' });
  expect(opened).toHaveLength(1);
  expect(opened[0].url).toMatch(/^chrome-extension:\/\/fixture\/save\/save.html\?draft=/);
  expect(opened[0].url).not.toContain('private');
  expect(Object.values(records)[0].data.code).toBe('private draft');
  expect(Object.values(records)[0].dialogTabId).toBe(2);
});

for (const surface of ['window', 'tab']) {
  test(`closing a save ${surface} while it opens removes its private draft`, async () => {
    let clicked, removed;
    const cleanups = [];
    const records = {};
    const messages = [];
    const chrome = {
      runtime: { onInstalled: { addListener() {} }, onMessage: { addListener() {} }, getURL: path => 'chrome-extension://fixture/' + path },
      contextMenus: { onClicked: { addListener(listener) { clicked = listener; } } },
      storage: {
        sync: { get: async () => ({ interactiveMode: true }) },
        session: {
          get: async () => structuredClone(records),
          set: async values => Object.assign(records, structuredClone(values)),
          remove: async key => { delete records[key]; }
        }
      },
      tabs: {
        sendMessage: async (_id, message) => {
          if (message.action === 'contextMenuSave') throw new Error('No receiver');
          messages.push(message);
        },
        onRemoved: { addListener(listener) { if (surface === 'tab') removed = listener; } },
        create: async () => { cleanups.push(removed(2)); return { id: 2 }; }
      }
    };
    if (surface === 'window') chrome.windows = {
      onRemoved: { addListener(listener) { removed = listener; } },
      create: async () => { cleanups.push(removed(2)); return { id: 2 }; }
    };
    const source = fs.readFileSync(require.resolve('../background.src.js'), 'utf8').replace(/^import .*;\n/, '');
    vm.runInNewContext(source, { chrome, console, URL, crypto: require('node:crypto').webcrypto, SnipoClient: class {} });
    await clicked({ menuItemId: 'save-to-snipo', selectionText: 'PRIVATE-DRAFT-MARKER' }, { id: 1, title: 'Example', url: 'https://code.test/' });
    await Promise.all(cleanups);
    expect(Object.keys(records)).toHaveLength(0);
    expect(messages.some(message => message.action === 'saveDialogClosed')).toBe(true);
  });
}
