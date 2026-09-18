import{r as e}from"./shader-type-decoder-C6G-bVH3.js";var t=`(?:var<\\s*(uniform|storage(?:\\s*,\\s*[A-Za-z_][A-Za-z0-9_]*)?)\\s*>|var)\\s+([A-Za-z_][A-Za-z0-9_]*)`,n=`\\s*`,r=[RegExp(`@binding\\(\\s*(auto|\\d+)\\s*\\)${n}@group\\(\\s*(\\d+)\\s*\\)${n}${t}`,`g`),RegExp(`@group\\(\\s*(\\d+)\\s*\\)${n}@binding\\(\\s*(auto|\\d+)\\s*\\)${n}${t}`,`g`)],i=[RegExp(`@binding\\(\\s*(auto|\\d+)\\s*\\)${n}@group\\(\\s*(\\d+)\\s*\\)${n}${t}`,`g`),RegExp(`@group\\(\\s*(\\d+)\\s*\\)${n}@binding\\(\\s*(auto|\\d+)\\s*\\)${n}${t}`,`g`)],a=[RegExp(`@binding\\(\\s*(\\d+)\\s*\\)${n}@group\\(\\s*(\\d+)\\s*\\)${n}${t}`,`g`),RegExp(`@group\\(\\s*(\\d+)\\s*\\)${n}@binding\\(\\s*(\\d+)\\s*\\)${n}${t}`,`g`)],o=[RegExp(`@binding\\(\\s*(auto)\\s*\\)\\s*@group\\(\\s*(\\d+)\\s*\\)\\s*${t}`,`g`),RegExp(`@group\\(\\s*(\\d+)\\s*\\)\\s*@binding\\(\\s*(auto)\\s*\\)\\s*${t}`,`g`),RegExp(`@binding\\(\\s*(auto)\\s*\\)\\s*@group\\(\\s*(\\d+)\\s*\\)(?:[\\s\\n\\r]*@[A-Za-z_][^\\n\\r]*)*[\\s\\n\\r]*${t}`,`g`),RegExp(`@group\\(\\s*(\\d+)\\s*\\)\\s*@binding\\(\\s*(auto)\\s*\\)(?:[\\s\\n\\r]*@[A-Za-z_][^\\n\\r]*)*[\\s\\n\\r]*${t}`,`g`)];function s(e){let t=e.split(``),n=0,r=0,i=!1,a=!1,o=!1;for(;n<e.length;){let s=e[n],c=e[n+1];if(a){o?o=!1:s===`\\`?o=!0:s===`"`&&(a=!1),n++;continue}if(i){s===`
`||s===`\r`?i=!1:t[n]=` `,n++;continue}if(r>0){if(s===`/`&&c===`*`){t[n]=` `,t[n+1]=` `,r++,n+=2;continue}if(s===`*`&&c===`/`){t[n]=` `,t[n+1]=` `,r--,n+=2;continue}s!==`
`&&s!==`\r`&&(t[n]=` `),n++;continue}if(s===`"`){a=!0,n++;continue}if(s===`/`&&c===`/`){t[n]=` `,t[n+1]=` `,i=!0,n+=2;continue}if(s===`/`&&c===`*`){t[n]=` `,t[n+1]=` `,r=1,n+=2;continue}n++}return t.join(``)}function c(e,t){let n=s(e),r=[];for(let i of t){i.lastIndex=0;let a;for(a=i.exec(n);a;){let o=i===t[0],s=a.index,c=a[0].length;r.push({match:e.slice(s,s+c),index:s,length:c,bindingToken:a[o?1:2],groupToken:a[o?2:1],accessDeclaration:a[3]?.trim(),name:a[4]}),a=i.exec(n)}}return r.sort((e,t)=>e.index-t.index)}function l(e,t,n){let r=c(e,t);if(!r.length)return e;let i=``,a=0;for(let t of r)i+=e.slice(a,t.index),i+=n(t),a=t.index+t.length;return i+=e.slice(a),i}function u(e){return/@binding\(\s*auto\s*\)/.test(s(e))}function d(e,t){return c(e,t===r||t===i?o:t).find(e=>e.bindingToken===`auto`)}function f(e,t={}){let n=p(e),r=m(n);if(!r)return null;let i=h(n,r);if(!i)return null;let a=_(n,r,i);if(!a)return null;if(t.scanVertexAttributes===!1)return{attributes:[],bindings:a};let o=g(n,r);if(!o)return null;let s=S(n,r,i,o,t.vertexEntryPoint);return s?{attributes:s,bindings:a}:null}function p(e){let t=s(e),n=/[A-Za-z_][A-Za-z0-9_]*|(?:0[xX][0-9A-Fa-f]+|\d+)|[@(){}<>\[\]:,;=]/g,r=[],i=n.exec(t);for(;i;)r.push({value:i[0],index:i.index}),i=n.exec(t);return r}function m(e){let t=[],n=0;for(let r of e){if(r.value===`}`&&n===0)return null;t.push(n),r.value===`{`?n++:r.value===`}`&&n--}return n===0?t:null}function h(e,t){let n=new Map;for(let r=0;r<e.length;r++){if(t[r]!==0||e[r].value!==`alias`)continue;let i=e[r+1]?.value;if(!j(i)||e[r+2]?.value!==`=`||n.has(i))return null;let a=ie(e,t,r+3,`;`);if(a<0||a===r+3)return null;n.set(i,A(e.slice(r+3,a))),r=a}return n}function g(e,t){let n=new Map;for(let r=0;r<e.length;r++){if(t[r]!==0||e[r].value!==`struct`)continue;let i=e[r+1]?.value,a=r+2;if(!j(i)||n.has(i)||e[a]?.value!==`{`)return null;let o=O(e,a,`{`,`}`);if(o<0)return null;n.set(i,e.slice(a+1,o)),r=o}return n}function _(e,t,n){let r=[],i=new Set,a=new Set;for(let o=0;o<e.length;o++){if(t[o]!==0||e[o].value!==`var`)continue;let s=ae(e,t,o),c=e.slice(s,o),l=te(c,`group`),u=te(c,`binding`);if(l===null||u===null||l===void 0!=(u===void 0))return null;if(l===void 0||u===void 0)continue;let d=o+1,f=[];if(e[d]?.value===`<`){let t=O(e,d,`<`,`>`);if(t<0)return null;let n=k(e.slice(d+1,t),`,`);if(!n)return null;f=n.map(A),d=t+1}let p=e[d]?.value;if(!j(p)||e[d+1]?.value!==`:`)return null;let m=ie(e,t,d+2,`;`);if(m<0||m===d+2)return null;let h=T(A(e.slice(d+2,m)),n);if(!h)return null;let g=v({name:p,group:l,location:u,addressSpace:f,resourceType:h}),_=`${l}:${u}`;if(!g||i.has(_)||a.has(p))return null;r.push(g),i.add(_),a.add(p),o=m}return x(r),r.sort((e,t)=>e.group-t.group||e.location-t.location||e.name.localeCompare(t.name))}function v(e){let{name:t,group:n,location:r,addressSpace:i,resourceType:a}=e,o={name:t,group:n,location:r};if(i[0]===`uniform`&&i.length===1)return{...o,type:`uniform`};if(i[0]===`storage`&&i.length<=2){let e=i[1]||`read`;return e===`read`?{...o,type:`read-only-storage`}:e===`read_write`?{...o,type:`storage`}:null}return i.length>0?null:a===`sampler`||a===`sampler_comparison`?{...o,type:`sampler`,...a===`sampler_comparison`?{samplerType:`comparison`}:{}}:a===`texture_external`?{...o,type:`external-texture`}:y(o,a)||b(o,a)}function y(e,t){let n=/^texture_storage_(1d|2d|2d_array|3d)<([A-Za-z0-9_]+),(read|write|read_write)>$/.exec(t);if(!n)return null;let r={read:`read-only`,write:`write-only`,read_write:`read-write`}[n[3]];return{...e,type:`storage`,format:n[2],access:r,viewDimension:D(n[1])}}function b(e,t){let n=/^texture_(multisampled_)?(1d|2d|2d_array|cube|cube_array|3d)<(f32|i32|u32)>$/.exec(t);if(n){if(n[1]&&n[2]!==`2d`)return null;let t={f32:`float`,i32:`sint`,u32:`uint`}[n[3]];return{...e,type:`texture`,viewDimension:D(n[2]),sampleType:t,multisampled:!!n[1]}}let r=/^texture_depth_(multisampled_)?(2d|2d_array|cube|cube_array)$/.exec(t);return!r||r[1]&&r[2]!==`2d`?null:{...e,type:`texture`,viewDimension:D(r[2]),sampleType:`depth`,multisampled:!!r[1]}}function x(e){for(let t of e){if(t.type!==`sampler`||t.samplerType||!t.name.endsWith(`Sampler`))continue;let n=t.name.slice(0,-7);e.find(e=>e.type===`texture`&&e.name===n&&e.group===t.group)?.sampleType===`depth`&&(t.samplerType=`non-filtering`)}}function S(e,t,n,r,i){let a=C(e,t);if(!a)return null;let o=a.filter(e=>e.vertex),s=i?o.find(e=>e.name===i):o.length===1?o[0]:void 0;if(!s)return o.length===0&&!i?[]:null;let c=k(s.parameters,`,`);if(!c)return null;let l=[],u=new Set,d=new Set,f=new Set;for(let e of c)if(e.length>0&&!w({declaration:e,aliases:n,structures:r,attributes:l,attributeLocations:u,attributeNames:d,visitedStructures:f}))return null;return l.sort((e,t)=>e.location-t.location||e.name.localeCompare(t.name))}function C(e,t){let n=[],r=new Set;for(let i=0;i<e.length;i++){if(t[i]!==0||e[i].value!==`fn`)continue;let a=e[i+1]?.value,o=i+2;if(!j(a)||r.has(a)||e[o]?.value!==`(`)return null;let s=O(e,o,`(`,`)`);if(s<0)return null;let c=ae(e,t,i);n.push({name:a,vertex:ne(e.slice(c,i),`vertex`),parameters:e.slice(o+1,s)}),r.add(a),i=s}return n}function w(e){let{declaration:t,aliases:n,structures:r,attributes:i,attributeLocations:a,attributeNames:o,visitedStructures:s}=e,c=re(t,`:`);if(c<1||c===t.length-1)return!1;let l=oe(t.slice(0,c)),u=te(t.slice(0,c),`location`),d=ne(t.slice(0,c),`builtin`),f=T(A(t.slice(c+1)),n);if(!l||u===null||!f||u!==void 0&&d)return!1;if(u!==void 0){let e=ee(f);return!e||a.has(u)||o.has(l)?!1:(i.push({name:l,location:u,type:e}),a.add(u),o.add(l),!0)}if(d)return!0;let p=r.get(f);if(!p||s.has(f))return!1;let m=k(p,`,`);if(!m)return!1;s.add(f);for(let t of m)if(t.length>0&&!w({...e,declaration:t}))return!1;return s.delete(f),!0}function T(e,t,n=new Set){let r=p(e),i=``;for(let e of r){let r=t.get(e.value);if(!r){i+=E(e.value);continue}if(n.has(e.value))return null;let a=new Set(n);a.add(e.value);let o=T(r,t,a);if(!o)return null;i+=o}return i}function E(e){let t=/^(vec[234]|mat[234]x[234])([fiuh])$/.exec(e);if(!t)return e;let n={f:`f32`,i:`i32`,u:`u32`,h:`f16`}[t[2]];return`${t[1]}<${n}>`}function ee(e){return/^(?:i32|u32|f32|f16|vec[234]<(?:i32|u32|f32|f16)>)$/.test(e)?e:null}function te(e,t){let n;for(let r=0;r<e.length;r++)if(e[r].value===`@`&&e[r+1]?.value===t){if(n!==void 0||e[r+2]?.value!==`(`||!/^\d+$/.test(e[r+3]?.value||``)||e[r+4]?.value!==`)`)return null;n=Number(e[r+3].value)}return n}function ne(e,t){return e.some((n,r)=>n.value===`@`&&e[r+1]?.value===t)}function D(e){return e.replace(`_`,`-`)}function O(e,t,n,r){let i=0;for(let a=t;a<e.length;a++)if(e[a].value===n)i++;else if(e[a].value===r&&--i===0)return a;return-1}function k(e,t){let n=[],r=0,i={"(":0,"<":0,"[":0,"{":0},a=Object.keys(i),o={")":`(`,">":`<`,"]":`[`,"}":`{`};for(let s=0;s<e.length;s++){let c=e[s].value;if(c===t&&a.every(e=>i[e]===0)){n.push(e.slice(r,s)),r=s+1;continue}if(c in i)i[c]++;else if(c in o){let e=o[c];if(i[e]--,i[e]<0)return null}}return a.every(e=>i[e]===0)?(n.push(e.slice(r)),n):null}function re(e,t){let n=k(e,t);return n&&n.length===2?n[0].length:-1}function ie(e,t,n,r){for(let i=n;i<e.length;i++)if(t[i]===0&&e[i].value===r)return i;return-1}function ae(e,t,n){for(let r=n-1;r>=0;r--)if(e[r].value===`;`&&t[r]===0||e[r].value===`}`&&t[r]===1)return r+1;return 0}function oe(e){for(let t=e.length-1;t>=0;t--)if(j(e[t].value))return e[t].value;return null}function A(e){return e.map(e=>e.value).join(``)}function j(e){return!!(e&&/^[A-Za-z_][A-Za-z0-9_]*$/.test(e))}function M(e,t){if(!e){let e=Error(t||`shadertools: assertion failed.`);throw Error.captureStackTrace?.(e,M),e}}var N={number:{type:`number`,validate(e,t){return Number.isFinite(e)&&typeof t==`object`&&(t.max===void 0||e<=t.max)&&(t.min===void 0||e>=t.min)}},array:{type:`array`,validate(e,t){return Array.isArray(e)||ArrayBuffer.isView(e)}}};function se(e){let t={};for(let[n,r]of Object.entries(e))t[n]=ce(r);return t}function ce(e){let t=le(e);if(t!==`object`)return{value:e,...N[t],type:t};if(typeof e==`object`)return e?e.type===void 0?e.value===void 0?{type:`object`,value:e}:(t=le(e.value),{...e,...N[t],type:t}):{...e,...N[e.type],type:e.type}:{type:`object`,value:null};throw Error(`props`)}function le(e){return Array.isArray(e)||ArrayBuffer.isView(e)?`array`:typeof e}var ue={vertex:`#ifdef MODULE_LOGDEPTH
  logdepth_adjustPosition(gl_Position);
#endif
`,fragment:`#ifdef MODULE_MATERIAL
  fragColor = material_filterColor(fragColor);
#endif

#ifdef MODULE_LIGHTING
  fragColor = lighting_filterColor(fragColor);
#endif

#ifdef MODULE_FOG
  fragColor = fog_filterColor(fragColor);
#endif

#ifdef MODULE_PICKING
  fragColor = picking_filterHighlightColor(fragColor);
  fragColor = picking_filterPickingColor(fragColor);
#endif

#ifdef MODULE_LOGDEPTH
  logdepth_setFragDepth();
#endif
`},de=/void\s+main\s*\([^)]*\)\s*\{\n?/,fe=/}\n?[^{}]*$/,pe=[],P=`__LUMA_INJECT_DECLARATIONS__`;function me(e){let t={vertex:{},fragment:{}};for(let n in e){let r=e[n],i=he(n);typeof r==`string`&&(r={order:0,injection:r}),t[i][n]=r}return t}function he(e){let t=e.slice(0,2);switch(t){case`vs`:return`vertex`;case`fs`:return`fragment`;default:throw Error(t)}}function F(e,t,n,r=!1,i=`glsl`,a={}){let o=t===`vertex`;for(let t in n){let r=n[t];r.sort((e,t)=>e.order-t.order),pe.length=r.length;for(let e=0,t=r.length;e<t;++e)pe[e]=r[e].injection;let s=`${pe.join(`
`)}\n`;switch(t){case`vs:#decl`:(i===`wgsl`||o)&&(e=e.replace(P,s));break;case`vs:#main-start`:(i===`wgsl`||o)&&(e=i===`wgsl`?I(e,`vertex`,s,`start`,a.vertex):e.replace(de,e=>e+s));break;case`vs:#main-end`:(i===`wgsl`||o)&&(e=i===`wgsl`?I(e,`vertex`,s,`end`,a.vertex):e.replace(fe,e=>s+e));break;case`fs:#decl`:(i===`wgsl`||!o)&&(e=e.replace(P,s));break;case`fs:#main-start`:(i===`wgsl`||!o)&&(e=i===`wgsl`?I(e,`fragment`,s,`start`,a.fragment):e.replace(de,e=>e+s));break;case`fs:#main-end`:(i===`wgsl`||!o)&&(e=i===`wgsl`?I(e,`fragment`,s,`end`,a.fragment):e.replace(fe,e=>s+e));break;default:e=e.replace(t,e=>e+s)}}return e=e.replace(P,``),r&&(e=e.replace(/\}\s*$/,e=>e+ue[t])),e}function I(e,t,n,r,i){let a=ge(e,t,i);if(!a)return e;if(r===`start`){let t=a.openBraceIndex+1;return`${e.slice(0,t)}\n${n}${e.slice(t)}`}return`${e.slice(0,a.closeBraceIndex)}${n}${e.slice(a.closeBraceIndex)}`}function ge(e,t,n){let r=t===`vertex`?`@vertex`:`@fragment`,i=e.indexOf(r);if(i<0)return null;let a=n?e.search(RegExp(`\\bfn\\s+${_e(n)}\\s*\\(`)):e.indexOf(`fn`,i);if(a<0)return null;let o=e.indexOf(`{`,a);if(o<0)return null;let s=0;for(let t=o;t<e.length;t++){let n=e[t];if(n===`{`)s++;else if(n===`}`&&(s--,s===0))return{openBraceIndex:o,closeBraceIndex:t}}return null}function _e(e){return e.replace(/[.*+?^${}()|[\]\\]/g,`\\$&`)}function L(e){e.map(e=>ve(e))}function ve(e){if(e.instance)return;L(e.dependencies||[]);let{propTypes:t={},deprecations:n=[],inject:r={}}=e,i={normalizedInjections:me(r),parsedDeprecations:be(n)};t&&(i.propValidators=se(t)),e.instance=i;let a={};t&&(a=Object.entries(t).reduce((e,[t,n])=>{let r=n?.value;return r&&(e[t]=r),e},{})),e.defaultUniforms={...e.defaultUniforms,...a}}function ye(e,t,n){e.deprecations?.forEach(e=>{e.regex?.test(t)&&(e.deprecated?n.deprecated(e.old,e.new)():n.removed(e.old,e.new)())})}function be(e){return e.forEach(e=>{switch(e.type){case`function`:e.regex=RegExp(`\\b${e.old}\\(`);break;default:e.regex=RegExp(`${e.type} ${e.old};`)}}),e}function R(e){L(e);let t={},n={};xe({modules:e,level:0,moduleMap:t,moduleDepth:n});let r=Object.keys(n).sort((e,t)=>n[t]-n[e]).map(e=>t[e]);return L(r),r}function xe(e){let{modules:t,level:n,moduleMap:r,moduleDepth:i}=e;if(n>=5)throw Error(`Possible loop in shader dependency graph`);for(let e of t)r[e.name]=e,(i[e.name]===void 0||i[e.name]<n)&&(i[e.name]=n);for(let e of t)e.dependencies&&xe({modules:e.dependencies,level:n+1,moduleMap:r,moduleDepth:i})}var Se=/^(?:uniform\s+)?(?:(?:lowp|mediump|highp)\s+)?[A-Za-z0-9_]+(?:<[^>]+>)?\s+([A-Za-z0-9_]+)(?:\s*\[[^\]]+\])?\s*;/,Ce=/((?:layout\s*\([^)]*\)\s*)*)uniform\s+([A-Za-z_][A-Za-z0-9_]*)\s*\{([\s\S]*?)\}\s*([A-Za-z_][A-Za-z0-9_]*)?\s*;/g;function z(e){return`${e.name}Uniforms`}function we(e,t){let n=t===`wgsl`?e.source:t===`vertex`?e.vs:e.fs;if(!n)return null;let r=z(e);return Oe(n,t===`wgsl`?`wgsl`:`glsl`,r)}function Te(e,t){let n=Object.keys(e.uniformTypes||{});if(!n.length)return null;let r=we(e,t);return r?{moduleName:e.name,uniformBlockName:z(e),stage:t,expectedUniformNames:n,actualUniformNames:r,matches:je(n,r)}:null}function Ee(e,t,n={}){let r=Te(e,t);if(!r||r.matches)return r;let i=Me(r);return n.log?.error?.(i,r)(),n.throwOnError!==!1&&M(!1,i),r}function B(e){let t=[],n=Ne(e);for(let e of n.matchAll(Ce)){let n=e[1]?.trim()||null;t.push({blockName:e[2],body:e[3],instanceName:e[4]||null,layoutQualifier:n,hasLayoutQualifier:!!n,isStd140:!!(n&&/\blayout\s*\([^)]*\bstd140\b[^)]*\)/.exec(n))})}return t}function De(e,t,n,r){let i=B(e).filter(e=>!e.isStd140),a=new Set;for(let e of i){if(a.has(e.blockName))continue;a.add(e.blockName);let i=r?.label?`${r.label} `:``,o=e.hasLayoutQualifier?`declares ${Pe(e.layoutQualifier)} instead of layout(std140)`:`does not declare layout(std140)`,s=`${i}${t} shader uniform block ${e.blockName} ${o}. luma.gl host-side shader block packing assumes explicit layout(std140) for GLSL uniform blocks. Add \`layout(std140)\` to the block declaration.`;n?.warn?.(s,e)()}return i}function Oe(e,t,n){let r=t===`wgsl`?ke(e,n):Ae(e,n);if(!r)return null;let i=[];for(let e of r.split(`
`)){let n=e.replace(/\/\/.*$/,``).trim();if(!n||n.startsWith(`#`))continue;let r=t===`wgsl`?n.match(/^([A-Za-z0-9_]+)\s*:/):n.match(Se);r&&i.push(r[1])}return i}function ke(e,t){let n=RegExp(`\\bstruct\\s+${t}\\b`,`m`).exec(e);if(!n)return null;let r=e.indexOf(`{`,n.index);if(r<0)return null;let i=0;for(let t=r;t<e.length;t++){let n=e[t];if(n===`{`){i++;continue}if(n===`}`&&(i--,i===0))return e.slice(r+1,t)}return null}function Ae(e,t){return B(e).find(e=>e.blockName===t)?.body||null}function je(e,t){if(e.length!==t.length)return!1;for(let n=0;n<e.length;n++)if(e[n]!==t[n])return!1;return!0}function Me(e){let{expectedUniformNames:t,actualUniformNames:n}=e,r=t.filter(e=>!n.includes(e)),i=n.filter(e=>!t.includes(e)),a=[`Expected ${t.length} fields, found ${n.length}.`],o=Fe(t,n);return o&&a.push(o),r.length&&a.push(`Missing from shader block (${r.length}): ${Ie(r)}.`),i.length&&a.push(`Unexpected in shader block (${i.length}): ${Ie(i)}.`),t.length<=12&&n.length<=12&&(r.length||i.length)&&(a.push(`Expected: ${t.join(`, `)}.`),a.push(`Actual: ${n.join(`, `)}.`)),`${e.moduleName}: ${e.stage} shader uniform block ${e.uniformBlockName} does not match module.uniformTypes. ${a.join(` `)}`}function Ne(e){return e.replace(/\/\*[\s\S]*?\*\//g,``).replace(/\/\/.*$/gm,``)}function Pe(e){return e.replace(/\s+/g,` `).trim()}function Fe(e,t){let n=Math.min(e.length,t.length);for(let r=0;r<n;r++)if(e[r]!==t[r])return`First mismatch at field ${r+1}: expected ${e[r]}, found ${t[r]}.`;return e.length>t.length?`Shader block ends after field ${t.length}; expected next field ${e[t.length]}.`:t.length>e.length?`Shader block has extra field ${t.length}: ${t[e.length]}.`:null}function Ie(e,t=8){if(e.length<=t)return e.join(`, `);let n=e.length-t;return`${e.slice(0,t).join(`, `)}, ... (${n} more)`}function Le(e){switch(e?.gpu.toLowerCase()){case`apple`:return`#define APPLE_GPU
// Apple optimizes away the calculation necessary for emulated fp64
#define LUMA_FP64_CODE_ELIMINATION_WORKAROUND 1
#define LUMA_FP32_TAN_PRECISION_WORKAROUND 1
// Intel GPU doesn't have full 32 bits precision in same cases, causes overflow
#define LUMA_FP64_HIGH_BITS_OVERFLOW_WORKAROUND 1
`;case`nvidia`:return`#define NVIDIA_GPU
// Nvidia optimizes away the calculation necessary for emulated fp64
#define LUMA_FP64_CODE_ELIMINATION_WORKAROUND 1
`;case`intel`:return`#define INTEL_GPU
// Intel optimizes away the calculation necessary for emulated fp64
#define LUMA_FP64_CODE_ELIMINATION_WORKAROUND 1
// Intel's built-in 'tan' function doesn't have acceptable precision
#define LUMA_FP32_TAN_PRECISION_WORKAROUND 1
// Intel GPU doesn't have full 32 bits precision in same cases, causes overflow
#define LUMA_FP64_HIGH_BITS_OVERFLOW_WORKAROUND 1
`;case`amd`:return`#define AMD_GPU
`;default:return`#define DEFAULT_GPU
// Prevent driver from optimizing away the calculation necessary for emulated fp64
#define LUMA_FP64_CODE_ELIMINATION_WORKAROUND 1
// Headless Chrome's software shader 'tan' function doesn't have acceptable precision
#define LUMA_FP32_TAN_PRECISION_WORKAROUND 1
// If the GPU doesn't have full 32 bits precision, will causes overflow
#define LUMA_FP64_HIGH_BITS_OVERFLOW_WORKAROUND 1
`}}function Re(e,t){if(Number(e.match(/^#version[ \t]+(\d+)/m)?.[1]||100)!==300)throw Error(`luma.gl v9 only supports GLSL 3.00 shader sources`);switch(t){case`vertex`:return e=He(e,Be),e;case`fragment`:return e=He(e,Ve),e;default:throw Error(t)}}var ze=[[/^(#version[ \t]+(100|300[ \t]+es))?[ \t]*\n/,`#version 300 es
`],[/\btexture(2D|2DProj|Cube)Lod(EXT)?\(/g,`textureLod(`],[/\btexture(2D|2DProj|Cube)(EXT)?\(/g,`texture(`]],Be=[...ze,[V(`attribute`),`in $1`],[V(`varying`),`out $1`]],Ve=[...ze,[V(`varying`),`in $1`]];function He(e,t){for(let[n,r]of t)e=e.replace(n,r);return e}function V(e){return RegExp(`\\b${e}[ \\t]+(\\w+[ \\t]+\\w+(\\[\\w+\\])?;)`,`g`)}function H(e,t,n=`glsl`){let r=``;for(let i in e){let a=e[i];if(r+=`${n===`wgsl`?`fn`:`void`} ${a.signature} {\n`,a.header&&(r+=`  ${a.header}`),t[i]){let e=t[i];e.sort((e,t)=>e.order-t.order);for(let t of e)r+=`  ${t.injection}\n`}a.footer&&(r+=`  ${a.footer}`),r+=`}
`}return r}function Ue(e){let t={vertex:{},fragment:{}};for(let n of e){let e,r;typeof n==`string`?(e={},r=n):(e=n,r=e.hook),r=r.trim();let i=r.indexOf(`:`),a=r.slice(0,i),o=r.slice(i+1),s=r.replace(/\(.+/,``),c=Object.assign(e,{signature:o});switch(a){case`vs`:t.vertex[s]=c;break;case`fs`:t.fragment[s]=c;break;default:throw Error(a)}}return t}function We(e,t){return{name:Ge(e,t),language:`glsl`,version:Ke(e)}}function Ge(e,t=`unnamed`){let n=/#define[^\S\r\n]*SHADER_NAME[^\S\r\n]*([A-Za-z0-9_-]+)\s*/.exec(e);return n?n[1]:t}function Ke(e){let t=100,n=e.match(/[^\s]+/g);if(n&&n.length>=2&&n[0]===`#version`){let e=parseInt(n[1],10);Number.isFinite(e)&&(t=e)}if(t!==100&&t!==300)throw Error(`Invalid GLSL version ${t}`);return t}var qe=[RegExp(`@binding\\(\\s*(\\d+)\\s*\\)\\s*@group\\(\\s*(\\d+)\\s*\\)\\s*${t}\\s*:\\s*([^;]+);`,`g`),RegExp(`@group\\(\\s*(\\d+)\\s*\\)\\s*@binding\\(\\s*(\\d+)\\s*\\)\\s*${t}\\s*:\\s*([^;]+);`,`g`)];function Je(e,t=[]){let n=s(e),r=new Map;for(let e of t)r.set(Xe(e.name,e.group,e.location),e.moduleName);let i=[];for(let e of qe){e.lastIndex=0;let t;for(t=e.exec(n);t;){let a=e===qe[0],o=Number(t[a?1:2]),s=Number(t[a?2:1]),c=t[3]?.trim(),l=t[4],u=t[5].trim(),d=r.get(Xe(l,s,o));i.push(Ye({name:l,group:s,binding:o,owner:d?`module`:`application`,moduleName:d,accessDeclaration:c,resourceType:u})),t=e.exec(n)}}return i.sort((e,t)=>e.group===t.group?e.binding===t.binding?e.name.localeCompare(t.name):e.binding-t.binding:e.group-t.group)}function Ye(e){let t={name:e.name,group:e.group,binding:e.binding,owner:e.owner,kind:`unknown`,moduleName:e.moduleName,resourceType:e.resourceType};if(e.accessDeclaration){let n=e.accessDeclaration.split(`,`).map(e=>e.trim());if(n[0]===`uniform`)return{...t,kind:`uniform`,access:`uniform`};if(n[0]===`storage`){let e=n[1]||`read_write`;return{...t,kind:e===`read`?`read-only-storage`:`storage`,access:e}}}return e.resourceType===`sampler`||e.resourceType===`sampler_comparison`?{...t,kind:`sampler`,samplerKind:e.resourceType===`sampler_comparison`?`comparison`:`filtering`}:e.resourceType.startsWith(`texture_storage_`)?{...t,kind:`storage-texture`,access:$e(e.resourceType),viewDimension:Ze(e.resourceType)}:e.resourceType.startsWith(`texture_`)?{...t,kind:`texture`,viewDimension:Ze(e.resourceType),sampleType:Qe(e.resourceType),multisampled:e.resourceType.startsWith(`texture_multisampled_`)}:t}function Xe(e,t,n){return`${t}:${n}:${e}`}function Ze(e){if(e.includes(`cube_array`))return`cube-array`;if(e.includes(`2d_array`))return`2d-array`;if(e.includes(`cube`))return`cube`;if(e.includes(`3d`))return`3d`;if(e.includes(`2d`))return`2d`;if(e.includes(`1d`))return`1d`}function Qe(e){if(e.startsWith(`texture_depth_`))return`depth`;if(e.includes(`<i32>`))return`sint`;if(e.includes(`<u32>`))return`uint`;if(e.includes(`<f32>`))return`float`}function $e(e){return/,\s*([A-Za-z_][A-Za-z0-9_]*)\s*>$/.exec(e)?.[1]}var U=`([a-zA-Z_][a-zA-Z0-9_]*)`,et=/^\s*\#\s*if\s+(.+?)\s*(?:\/\/.*)?$/,tt=RegExp(`^\\s*\\#\\s*ifdef\\s*${U}\\s*$`),nt=RegExp(`^\\s*\\#\\s*ifndef\\s*${U}\\s*(?:\\/\\/.*)?$`),rt=/^\s*\#\s*else\s*(?:\/\/.*)?$/,it=/^\s*\#\s*endif\s*$/,at=RegExp(`^\\s*\\#\\s*ifdef\\s*${U}\\s*(?:\\/\\/.*)?$`),ot=/^\s*\#\s*endif\s*(?:\/\/.*)?$/;function W(e,t){let n=e.split(`
`),r=[],i=[],a=!0;for(let e of n){let n=e.match(et),o=e.match(at)||e.match(tt),s=e.match(nt),c=e.match(rt),l=e.match(ot)||e.match(it);if(n){let e=st(n[1],t?.defines||{}),r=a&&e;i.push({parentActive:a,branchTaken:e,active:r}),a=r}else if(o||s){let e=(o||s)?.[1],n=!!t?.defines?.[e],r=o?n:!n,c=a&&r;i.push({parentActive:a,branchTaken:r,active:c}),a=c}else if(c){let e=i[i.length-1];if(!e)throw Error(`Encountered #else without matching #if, #ifdef or #ifndef`);e.active=e.parentActive&&!e.branchTaken,e.branchTaken=!0,a=e.active}else l?(i.pop(),a=!i.length||i[i.length-1].active):a&&r.push(e)}if(i.length>0)throw Error(`Unterminated conditional block in shader source`);return r.join(`
`)}function st(e,t){let n=e.trim();if(/^[+-]?\d+(?:\.\d+)?$/.test(n))return Number(n)!==0;if(n===`true`)return!0;if(n===`false`)return!1;let r=n.match(RegExp(`^!\\s*${U}$`));if(r)return!t[r[1]];let i=n.match(RegExp(`^${U}$`));if(i)return!!t[i[1]];let a=n.match(RegExp(`^defined\\s*\\(\\s*${U}\\s*\\)$`));if(a)return t[a[1]]!==void 0;let o=n.match(RegExp(`^!\\s*defined\\s*\\(\\s*${U}\\s*\\)$`));if(o)return t[o[1]]===void 0;throw Error(`Unsupported #if expression "${e}"`)}function ct(e,t){let n=[];for(let[r,i]of Object.entries(t))ut(e,r),n.push(`in ${G(i)} ${r};`);return n.join(`
`)}function lt(e,t,n){let r=Object.entries(n);if(r.length===0)return{source:e,declarations:``,initialization:``};let i=dt(e,t),a=e.slice(i.openParenthesis+1,i.closeParenthesis),o=ft(e,a),s=new Set(o.locations),c=[],l=[],u=[];for(let[t,n]of r){if(o.names.has(t)||_t(e,t))throw Error(`ShaderPlugin vertex input "${t}" conflicts with an existing WGSL shader input or variable`);let r=vt(s);s.add(r);let i=`_luma_${t}`;c.push(`@location(${r}) ${i}: ${n}`),l.push(`var<private> ${t}: ${n};`),u.push(`${t} = ${i};`)}let d=a.trim()?`,
  `:`
  `,f=a.trim()?``:`
`,p=`${a}${d}${c.join(`,
  `)}${f}`;return{source:e.slice(0,i.openParenthesis+1)+p+e.slice(i.closeParenthesis),declarations:l.join(`
`),initialization:u.join(`
`)}}function G(t){let{primitiveType:n,components:r}=e.getAttributeShaderTypeInfo(t),i=n===`i32`?`int`:n===`u32`?`uint`:`float`;return r===1?i:`${i===`int`?`i`:i===`uint`?`u`:``}vec${r}`}function ut(e,t){let n=K(t);if(RegExp(`\\b(?:in|attribute)\\s+(?:(?:lowp|mediump|highp)\\s+)?[A-Za-z_][A-Za-z0-9_]*\\s+${n}\\s*(?:\\[|;)`).test(e))throw Error(`ShaderPlugin vertex input "${t}" conflicts with an existing GLSL input`)}function dt(e,t){let n=RegExp(`\\bfn\\s+${K(t)}\\s*\\(`,`g`).exec(e);if(!n)throw Error(`ShaderPlugin vertex inputs require WGSL vertex entry point "${t}"`);let r=e.indexOf(`(`,n.index),i=yt(e,r,`(`,`)`);if(i<0)throw Error(`Unable to parse WGSL vertex entry point "${t}" parameters`);return{openParenthesis:r,closeParenthesis:i}}function ft(e,t){let n=pt(t),r=new Set(mt(t)),i=ht(t);for(let t of i){let i=gt(e,t);if(i!==null){n.push(...pt(i));for(let e of mt(i))r.add(e)}}return{locations:n,names:r}}function pt(e){let t=[],n=/@location\s*\(\s*(\d+)\s*\)/g,r=n.exec(e);for(;r;)t.push(Number(r[1])),r=n.exec(e);return t}function mt(e){let t=[],n=/(?:^|,)\s*(?:@[A-Za-z_][\w]*(?:\([^)]*\))?\s*)*([A-Za-z_][\w]*)\s*:/gm,r=n.exec(e);for(;r;)t.push(r[1]),r=n.exec(e);return t}function ht(e){let t=[],n=/:\s*([A-Za-z_][\w]*)\b/g,r=n.exec(e);for(;r;)t.push(r[1]),r=n.exec(e);return t}function gt(e,t){let n=RegExp(`\\bstruct\\s+${K(t)}\\s*\\{`,`g`).exec(e);if(!n)return null;let r=e.indexOf(`{`,n.index),i=yt(e,r,`{`,`}`);return i<0?null:e.slice(r+1,i)}function _t(e,t){let n=K(t),r=RegExp(`\\b(?:var(?:<[^>]+>)?|let|const)\\s+${n}\\b`,`g`),i=r.exec(e);for(;i;){if(bt(e,i.index)===0)return!0;i=r.exec(e)}return!1}function vt(e){let t=0;for(;e.has(t);)t++;return t}function yt(e,t,n,r){let i=0,a=0,o=!1;for(let s=t;s<e.length;s++){let t=e[s],c=e[s+1];if(o){t===`
`&&(o=!1);continue}if(a>0){t===`/`&&c===`*`?(a++,s++):t===`*`&&c===`/`&&(a--,s++);continue}if(t===`/`&&c===`/`){o=!0,s++;continue}if(t===`/`&&c===`*`){a=1,s++;continue}if(t===n&&i++,t===r&&--i===0)return s}return-1}function bt(e,t){let n=0,r=0,i=!1;for(let a=0;a<t;a++){let t=e[a],o=e[a+1];if(i){t===`
`&&(i=!1);continue}if(r>0){t===`/`&&o===`*`?(r++,a++):t===`*`&&o===`/`&&(r--,a++);continue}t===`/`&&o===`/`?(i=!0,a++):t===`/`&&o===`*`?(r=1,a++):t===`{`?n++:t===`}`&&n--}return n}function K(e){return e.replace(/[.*+?^${}()|[\]\\]/g,`\\$&`)}function xt(e,t,n){let r=[],i=[];for(let[a,o]of Object.entries(n)){zt(e,a);let n=o.interpolation===`flat`?`flat `:``,s=t===`vertex`?`out`:`in`;r.push(`${n}${s} ${G(o.type)} ${a};`),t===`vertex`&&i.push(`${a} = ${Lt(o.type)};`)}return{declarations:r.join(`
`),initialization:i.join(`
`)}}function St(e,t,n,r){let i=Object.entries(r);if(i.length===0)return{source:e,declarations:``,vertexInitialization:``,fragmentInitialization:``};let a=e,o=q(a,t,`vertex`),s=Ct(a,o),c=q(a,n,`fragment`),l=wt(a,c),u=Tt(a,s),d=Tt(a,l.type),f=new Set([...J(o.parameters),...J(u.body),...J(c.parameters),...J(d.body)]),p=new Set([...Pt(u.body),...Pt(d.body)]),m=[],h=[],g=[],_=[];for(let[e,t]of i){if(f.has(e)||Ft(a,e))throw Error(`ShaderPlugin varying "${e}" conflicts with existing WGSL stage I/O or a module variable`);let n=It(p);p.add(n);let r=t.interpolation===`flat`?` @interpolate(flat)`:``;m.push(`  @location(${n})${r} ${e}: ${t.type},`),h.push(`var<private> ${e}: ${t.type};`),g.push(`${e} = ${Rt(t.type)};`),_.push(`${e} = ${l.name}.${e};`)}Dt(a,s,o.openBrace,o.closeBrace),a=Ot(a,s,o,i.map(([e])=>e)),o=q(a,t,`vertex`),a=kt(a,o,i.map(([e])=>e));let v=(s===l.type?[s]:[s,l.type]).map(e=>Tt(a,e).closeBrace).sort((e,t)=>t-e);for(let e of v)a=a.slice(0,e)+`${m.join(`
`)}\n`+a.slice(e);if(c=q(a,n,`fragment`),!RegExp(`\\b${X(l.name)}\\s*:`).test(c.parameters))throw Error(`Unable to preserve WGSL fragment input "${l.name}"`);return{source:a,declarations:h.join(`
`),vertexInitialization:g.join(`
`),fragmentInitialization:_.join(`
`)}}function q(e,t,n){let r=RegExp(`\\bfn\\s+${X(t)}\\s*\\(`,`g`).exec(e);if(!r)throw Error(`ShaderPlugin varyings require WGSL ${n} entry point "${t}"`);let i=e.indexOf(`(`,r.index),a=Y(e,i,`(`,`)`),o=e.indexOf(`{`,a),s=Y(e,o,`{`,`}`);if(a<0||o<0||s<0)throw Error(`Unable to parse WGSL ${n} entry point "${t}"`);return{openParenthesis:i,closeParenthesis:a,openBrace:o,closeBrace:s,parameters:e.slice(i+1,a)}}function Ct(e,t){let n=e.slice(t.closeParenthesis+1,t.openBrace),r=/->\s*([A-Za-z_][\w]*)\s*$/.exec(n.trim());if(!r||Et(e,r[1])===null)throw Error(`ShaderPlugin varyings require the WGSL vertex entry point to return a named struct`);return r[1]}function wt(e,t){let n=[];for(let r of Nt(t.parameters,`,`)){let t=/(?:@[A-Za-z_][\w]*(?:\([^)]*\))?\s*)*([A-Za-z_][\w]*)\s*:\s*([A-Za-z_][\w]*)\s*$/.exec(r.trim());t&&Et(e,t[2])&&n.push({name:t[1],type:t[2]})}if(n.length!==1)throw Error(`ShaderPlugin varyings require exactly one named WGSL fragment input struct; found ${n.length}`);return n[0]}function Tt(e,t){let n=Et(e,t);if(!n)throw Error(`Unable to find WGSL stage I/O struct "${t}"`);return n}function Et(e,t){let n=RegExp(`\\bstruct\\s+${X(t)}\\s*\\{`,`g`).exec(e);if(!n)return null;let r=e.indexOf(`{`,n.index),i=Y(e,r,`{`,`}`);return i<0?null:{openBrace:r,closeBrace:i,body:e.slice(r+1,i)}}function Dt(e,t,n,r){let i=RegExp(`\\b${X(t)}\\s*\\(`,`g`),a=i.exec(e);for(;a;){if(a.index<n||a.index>r)throw Error(`ShaderPlugin varying output struct "${t}" is constructed outside the selected vertex entry point`);a=i.exec(e)}}function Ot(e,t,n,r){let i=RegExp(`\\b${X(t)}\\s*\\(`,`g`),a=[],o=i.exec(e);for(;o;){if(o.index>n.openBrace&&o.index<n.closeBrace){let r=e.indexOf(`(`,o.index),i=Y(e,r,`(`,`)`);if(i<0||i>n.closeBrace)throw Error(`Unable to parse WGSL output constructor "${t}"`);a.push({openParenthesis:r,closeParenthesis:i})}o=i.exec(e)}for(let t of a.sort((e,t)=>t.closeParenthesis-e.closeParenthesis)){let n=e.slice(t.openParenthesis+1,t.closeParenthesis).trim()?`, `:``;e=e.slice(0,t.closeParenthesis)+n+r.join(`, `)+e.slice(t.closeParenthesis)}return e}function kt(e,t,n){let r=At(e,t.openBrace+1,t.closeBrace);for(let t=r.length-1;t>=0;t--){let i=r[t],a=e.slice(i.expressionStart,i.semicolon).trim();if(!a)throw Error(`ShaderPlugin varying vertex entry point cannot use an empty return`);let o=`_luma_vertexOutput${t}`,s=`{\nvar ${o} = ${a};\n${n.map(e=>`${o}.${e} = ${e};`).join(`
`)}\nreturn ${o};\n}`;e=e.slice(0,i.start)+s+e.slice(i.semicolon+1)}return e}function At(e,t,n){let r=[],i=t;for(;i<n;)if(i=Mt(e,i,n),e.slice(i,i+6)===`return`&&!/[A-Za-z0-9_]/.test(e[i+6]||``)){let t=i+6,a=jt(e,t,n);if(a<0)throw Error(`Unable to parse WGSL return statement in selected vertex entry point`);r.push({start:i,expressionStart:t,semicolon:a}),i=a+1}else i++;return r}function jt(e,t,n){let r=0,i=0;for(let a=t;a<n;a++){let t=Mt(e,a,n);if(t!==a){a=t-1;continue}let o=e[a];if(o===`(`&&r++,o===`)`&&r--,o===`[`&&i++,o===`]`&&i--,o===`;`&&r===0&&i===0)return a}return-1}function Mt(e,t,n){let r=t;if(e[r]===`/`&&e[r+1]===`/`){let t=e.indexOf(`
`,r+2);return t<0||t>n?n:t+1}if(e[r]===`/`&&e[r+1]===`*`){let t=1;for(r+=2;r<n&&t>0;)e[r]===`/`&&e[r+1]===`*`?(t++,r+=2):e[r]===`*`&&e[r+1]===`/`?(t--,r+=2):r++}return r}function Nt(e,t){let n=[],r=0,i=0,a=0;for(let o=0;o<e.length;o++){let s=e[o];s===`(`&&i++,s===`)`&&i--,s===`<`&&a++,s===`>`&&a--,s===t&&i===0&&a===0&&(n.push(e.slice(r,o)),r=o+1)}return n.push(e.slice(r)),n}function Pt(e){let t=[],n=/@location\s*\(\s*(\d+)\s*\)/g,r=n.exec(e);for(;r;)t.push(Number(r[1])),r=n.exec(e);return t}function J(e){let t=[],n=/(?:^|,)\s*(?:@[A-Za-z_][\w]*(?:\([^)]*\))?\s*)*([A-Za-z_][\w]*)\s*:/gm,r=n.exec(e);for(;r;)t.push(r[1]),r=n.exec(e);return t}function Ft(e,t){let n=RegExp(`\\b(?:var(?:<[^>]+>)?|let|const)\\s+${X(t)}\\b`,`g`),r=n.exec(e);for(;r;){if(Bt(e,r.index)===0)return!0;r=n.exec(e)}return!1}function It(e){let t=0;for(;e.has(t);)t++;return t}function Lt(t){let{primitiveType:n,components:r}=e.getAttributeShaderTypeInfo(t),i=n===`u32`?`0u`:n===`i32`?`0`:`0.0`;return r===1?i:`${G(t)}(${i})`}function Rt(t){let{primitiveType:n,components:r}=e.getAttributeShaderTypeInfo(t),i=`${n}(0)`;return r===1?i:`${t}(${i})`}function zt(e,t){if(RegExp(`\\b(?:flat\\s+|smooth\\s+)?(?:in|out|varying)\\s+(?:(?:lowp|mediump|highp)\\s+)?[A-Za-z_][A-Za-z0-9_]*\\s+${X(t)}\\s*(?:\\[|;)`).test(e))throw Error(`ShaderPlugin varying "${t}" conflicts with existing GLSL stage I/O`)}function Y(e,t,n,r){let i=0,a=0,o=!1;for(let s=t;s<e.length;s++){let t=e[s],c=e[s+1];if(o){t===`
`&&(o=!1);continue}if(a>0){t===`/`&&c===`*`?(a++,s++):t===`*`&&c===`/`&&(a--,s++);continue}if(t===`/`&&c===`/`){o=!0,s++;continue}if(t===`/`&&c===`*`){a=1,s++;continue}if(t===n&&i++,t===r&&--i===0)return s}return-1}function Bt(e,t){let n=0;for(let r=0;r<t;r++){let i=Mt(e,r,t);if(i!==r){r=i-1;continue}e[r]===`{`&&n++,e[r]===`}`&&n--}return n}function X(e){return e.replace(/[.*+?^${}()|[\]\\]/g,`\\$&`)}var Vt=`\n\n${P}\n`,Z=100,Ht=`precision highp float;
`;function Ut(e){let t=R(e.modules||[]),{source:n,bindingAssignments:r}=Gt(e.platformInfo,{...e,source:e.source,stage:`vertex`,modules:t});return{source:n,getUniforms:qt(t),bindingAssignments:r,bindingTable:Je(n,r),shaderLayout:f(n,{vertexEntryPoint:e.vertexEntryPoint,scanVertexAttributes:e.scanVertexAttributes})}}function Wt(e){let{vs:t,fs:n}=e,r=R(e.modules||[]);return{vs:Kt(e.platformInfo,{...e,source:t,stage:`vertex`,modules:r}),fs:Kt(e.platformInfo,{...e,source:n,stage:`fragment`,modules:r}),getUniforms:qt(r)}}function Gt(e,t){let{source:n,stage:r,modules:i,defines:a={},hookFunctions:o=[],inject:s={},pluginInjections:c={},pluginVertexInputs:l={},pluginVaryings:u={},vertexEntryPoint:d=`vertexMain`,fragmentEntryPoint:f=`fragmentMain`,log:p}=t;M(typeof n==`string`,`shader source must be a string`);let m=lt(W(n,{defines:a}),d,l),h=St(m.source,d,f,u),g=h.source,_=``,v=Ue(o),y={},b={},x={};Jt(c,y,b,x);for(let e in s){let t=typeof s[e]==`string`?{injection:s[e],order:0}:s[e],n=/^(v|f)s:(#)?([\w-]+)$/.exec(e);if(n){let r=n[2],i=n[3];r?i===`decl`?b[e]=[t]:x[e]=[t]:y[e]=[t]}else x[e]=[t]}Yt(m.declarations,m.initialization,b,x),Xt(h,b,x);let S=i,C=rn(g),w=nn(C.source),T=cn(S,t._bindingRegistry,w,a),E=[];for(let e of S){p&&ye(e,g,p);let n=an(W(tn(e,`wgsl`,p),{defines:a}),e,{usedBindingsByGroup:w,bindingRegistry:t._bindingRegistry,reservedBindingKeysByGroup:T});E.push(...n.bindingAssignments);let r=n.source;_+=r;let i=Zt(e);for(let e in i){let t=/^(v|f)s:#([\w-]+)$/.exec(e);if(t){let n=t[2]===`decl`?b:x;n[e]=n[e]||[],n[e].push(i[e])}else y[e]=y[e]||[],y[e].push(i[e])}}return _+=Vt,_=F(_,r,Qt(b),!1,`wgsl`,{vertex:d,fragment:f}),_+=$t(v,y),_+=hn(E),_+=C.source,_=F(_,r,x,!1,`wgsl`,{vertex:d,fragment:f}),mn(_),{source:_,bindingAssignments:E}}function Kt(e,t){let{source:n,stage:r,language:i=`glsl`,modules:a,defines:o={},hookFunctions:s=[],inject:c={},pluginInjections:l={},pluginVertexInputs:u={},pluginVaryings:d={},prologue:f=!0,log:p}=t;M(typeof n==`string`,`shader source must be a string`);let m=i===`glsl`?We(n).version:-1,h=e.shaderLanguageVersion,g=m===100?`#version 100`:`#version 300 es`,_=n.split(`
`).slice(1).join(`
`),v={};a.forEach(e=>{Object.assign(v,e.defines)}),Object.assign(v,o);let y=``;switch(i){case`wgsl`:break;case`glsl`:y=f?`\
${g}

// ----- PROLOGUE -------------------------
${`#define SHADER_TYPE_${r.toUpperCase()}`}

${Le(e)}
${r===`fragment`?Ht:``}

// ----- APPLICATION DEFINES -------------------------

${en(v)}

`:`${g}
`}let b=Ue(s),x={},S={},C={};Jt(l,x,S,C);for(let e in c){let t=typeof c[e]==`string`?{injection:c[e],order:0}:c[e],n=/^(v|f)s:(#)?([\w-]+)$/.exec(e);if(n){let r=n[2],i=n[3];r?i===`decl`?S[e]=[t]:C[e]=[t]:x[e]=[t]}else C[e]=[t]}if(r===`vertex`){let e=ct(_,u);e&&(S[`vs:#decl`]=S[`vs:#decl`]||[],S[`vs:#decl`].push({injection:e,order:-(2**53-1)}))}let w=xt(_,r,d);if(w.declarations){let e=r===`vertex`?`vs:#decl`:`fs:#decl`;S[e]=S[e]||[],S[e].push({injection:w.declarations,order:-(2**53-1)})}w.initialization&&(C[`vs:#main-start`]=C[`vs:#main-start`]||[],C[`vs:#main-start`].push({injection:w.initialization,order:-(2**53-1)}));for(let e of a){p&&ye(e,_,p);let t=tn(e,r,p);y+=t;let n=e.instance?.normalizedInjections[r]||{};for(let e in n){let t=/^(v|f)s:#([\w-]+)$/.exec(e);if(t){let r=t[2]===`decl`?S:C;r[e]=r[e]||[],r[e].push(n[e])}else x[e]=x[e]||[],x[e].push(n[e])}}return y+=`// ----- MAIN SHADER SOURCE -------------------------`,y+=Vt,y=F(y,r,S),y+=H(b[r],x),y+=_,y=F(y,r,C),i===`glsl`&&m!==h&&(y=Re(y,r)),i===`glsl`&&De(y,r,p),y.trim()}function qt(e){return function(t){let n={};for(let r of e){let e=r.getUniforms?.(t,n);Object.assign(n,e)}return n}}function Jt(e,t,n,r){for(let i in e){let a=/^(v|f)s:(#)?([\w-]+)$/.exec(i);if(a){let o=a[2],s=a[3],c=o?s===`decl`?n:r:t;c[i]=c[i]||[],c[i].push(...e[i])}else r[i]=r[i]||[],r[i].push(...e[i])}}function Yt(e,t,n,r){e&&(n[`vs:#decl`]=n[`vs:#decl`]||[],n[`vs:#decl`].push({injection:e,order:-(2**53-1)})),t&&(r[`vs:#main-start`]=r[`vs:#main-start`]||[],r[`vs:#main-start`].push({injection:t,order:-(2**53-1)}))}function Xt(e,t,n){e.declarations&&(t[`vs:#decl`]=t[`vs:#decl`]||[],t[`vs:#decl`].push({injection:e.declarations,order:-(2**53-1)})),e.vertexInitialization&&(n[`vs:#main-start`]=n[`vs:#main-start`]||[],n[`vs:#main-start`].push({injection:e.vertexInitialization,order:-(2**53-1)})),e.fragmentInitialization&&(n[`fs:#main-start`]=n[`fs:#main-start`]||[],n[`fs:#main-start`].push({injection:e.fragmentInitialization,order:-(2**53-1)}))}function Zt(e){return{...e.instance?.normalizedInjections.vertex||{},...e.instance?.normalizedInjections.fragment||{}}}function Qt(e){let t=[...e[`vs:#decl`]||[],...e[`fs:#decl`]||[]];return t.length?{"vs:#decl":t}:{}}function $t(e,t){return H(e.vertex,t,`wgsl`)+H(e.fragment,t,`wgsl`)}function en(e={}){let t=``;for(let n in e){let r=e[n];(r||Number.isFinite(r))&&(t+=`#define ${n.toUpperCase()} ${e[n]}\n`)}return t}function tn(e,t,n){let r;switch(t){case`vertex`:r=e.vs||``;break;case`fragment`:r=e.fs||``;break;case`wgsl`:r=e.source||``;break;default:M(!1)}if(!e.name)throw Error(`Shader module must have a name`);Ee(e,t,{log:n});let i=e.name.toUpperCase().replace(/[^0-9a-z]/gi,`_`),a=`\
// ----- MODULE ${e.name} ---------------

`;return t!==`wgsl`&&(a+=`#define MODULE_${i}\n`),a+=`${r}\n`,a}function nn(e){let t=new Map;for(let n of c(e,a)){let e=Number(n.bindingToken),r=Number(n.groupToken);Q(r,e,n.name),$(t,r,e,`application binding "${n.name}"`)}return t}function rn(e){let t=c(e,i),n=new Map;for(let e of t){if(e.bindingToken===`auto`)continue;let t=Number(e.bindingToken),r=Number(e.groupToken);Q(r,t,e.name),$(n,r,t,`application binding "${e.name}"`)}let r={sawSupportedBindingDeclaration:t.length>0},a=l(e,i,e=>sn(e,n,r));if(u(e)&&!r.sawSupportedBindingDeclaration)throw Error(`Unsupported @binding(auto) declaration form in application WGSL. Use adjacent "@group(N)" and "@binding(auto)" decorators followed by a bindable "var" declaration.`);return{source:a}}function an(e,t,n){let i=[],a={sawSupportedBindingDeclaration:c(e,r).length>0,nextHintedBindingLocation:typeof t.firstBindingSlot==`number`?t.firstBindingSlot:null},o=l(e,r,e=>on(e,{module:t,context:n,bindingAssignments:i,relocationState:a}));if(u(e)&&!a.sawSupportedBindingDeclaration)throw Error(`Unsupported @binding(auto) declaration form in module "${t.name}". Use adjacent "@group(N)" and "@binding(auto)" decorators followed by a bindable "var" declaration.`);return{source:o,bindingAssignments:i}}function on(e,t){let{module:n,context:r,bindingAssignments:i,relocationState:a}=t,{match:o,bindingToken:s,groupToken:c,name:l}=e,u=Number(c);if(s===`auto`){let e=gn(u,n.name,l),t=r.bindingRegistry?.get(e),s=t===void 0?fn(u,r.usedBindingsByGroup,n.name,a.nextHintedBindingLocation??void 0,r.bindingRegistry):t;return dn(n.name,u,s,l),t!==void 0&&ln(r.reservedBindingKeysByGroup,u,s,e)?(i.push({moduleName:n.name,name:l,group:u,location:s}),o.replace(/@binding\(\s*auto\s*\)/,`@binding(${s})`)):($(r.usedBindingsByGroup,u,s,`module "${n.name}" binding "${l}"`),r.bindingRegistry?.set(e,s),i.push({moduleName:n.name,name:l,group:u,location:s}),a.nextHintedBindingLocation!==null&&t===void 0&&(a.nextHintedBindingLocation=s+1),o.replace(/@binding\(\s*auto\s*\)/,`@binding(${s})`))}let d=Number(s);return dn(n.name,u,d,l),$(r.usedBindingsByGroup,u,d,`module "${n.name}" binding "${l}"`),i.push({moduleName:n.name,name:l,group:u,location:d}),o}function sn(e,t,n){let{match:r,bindingToken:i,groupToken:a,name:o}=e,s=Number(a);if(i===`auto`){let e=pn(s,t);return Q(s,e,o),$(t,s,e,`application binding "${o}"`),r.replace(/@binding\(\s*auto\s*\)/,`@binding(${e})`)}return n.sawSupportedBindingDeclaration=!0,r}function cn(e,t,n,r){let i=new Map;if(!t)return i;for(let a of e)for(let e of un(a,r)){let r=gn(e.group,a.name,e.name),o=t.get(r);if(o!==void 0){let t=i.get(e.group)||new Map,a=t.get(o);if(a&&a!==r)throw Error(`Duplicate WGSL binding reservation for modules "${a}" and "${r}": group ${e.group}, binding ${o}.`);$(n,e.group,o,`registered module binding "${r}"`),t.set(o,r),i.set(e.group,t)}}return i}function ln(e,t,n,r){let i=e.get(t);if(!i)return!1;let a=i.get(n);if(!a)return!1;if(a!==r)throw Error(`Registered module binding "${r}" collided with "${a}": group ${t}, binding ${n}.`);return!0}function un(e,t){let n=[],i=W(e.source||``,{defines:t});for(let e of c(i,r))n.push({name:e.name,group:Number(e.groupToken)});return n}function Q(e,t,n){if(e===0&&t>=Z)throw Error(`Application binding "${n}" in group 0 uses reserved binding ${t}. Application-owned explicit group-0 bindings must stay below ${Z}.`)}function dn(e,t,n,r){if(t===0&&n<Z)throw Error(`Module "${e}" binding "${r}" in group 0 uses reserved application binding ${n}. Module-owned explicit group-0 bindings must be ${Z} or higher.`)}function $(e,t,n,r){let i=e.get(t)||new Set;if(i.has(n))throw Error(`Duplicate WGSL binding assignment for ${r}: group ${t}, binding ${n}.`);i.add(n),e.set(t,i)}function fn(e,t,n,r,i){let a=t.get(e)||new Set,o=new Set,s=`${e}:`,c=`${s}${n}:`;for(let[e,t]of i||[])e.startsWith(c)&&o.add(t);let l=r??(e===0?Z:a.size>0?Math.max(...a)+1:0);for(;a.has(l)||o.has(l);)l++;for(let[e,t]of i||[])t===l&&e.startsWith(s)&&i?.delete(e);return l}function pn(e,t){let n=t.get(e)||new Set,r=0;for(;n.has(r);)r++;return r}function mn(e){let t=d(e,r);if(!t)return;let n=_n(e,t.index);throw n?Error(`Unresolved @binding(auto) for module "${n}" binding "${t.name}" remained in assembled WGSL source.`):vn(e,t.index)?Error(`Unresolved @binding(auto) for application binding "${t.name}" remained in assembled WGSL source.`):Error(`Unresolved @binding(auto) remained in assembled WGSL source near "${yn(t.match)}".`)}function hn(e){if(e.length===0)return``;let t=`// ----- MODULE WGSL BINDING ASSIGNMENTS ---------------
`;for(let n of e)t+=`// ${n.moduleName}.${n.name} -> @group(${n.group}) @binding(${n.location})\n`;return t+=`
`,t}function gn(e,t,n){return`${e}:${t}:${n}`}function _n(e,t){let n=/^\/\/ ----- MODULE ([^\n]+) ---------------$/gm,r,i;for(i=n.exec(e);i&&i.index<=t;)r=i[1],i=n.exec(e);return r}function vn(e,t){let n=e.indexOf(Vt);return n>=0?t>n:!0}function yn(e){return e.replace(/\s+/g,` `).trim()}var bn=class e{static defaultShaderAssemblers={};_hookFunctions=[];_defaultModules=[];static getDefaultShaderAssembler(t){return M(t===`glsl`||t===`wgsl`),t===`wgsl`?(e.defaultShaderAssemblers.wgsl=e.defaultShaderAssemblers.wgsl||new Sn,e.defaultShaderAssemblers.wgsl):(e.defaultShaderAssemblers.glsl=e.defaultShaderAssemblers.glsl||new xn,e.defaultShaderAssemblers.glsl)}addDefaultModule(e){this._defaultModules.find(t=>t.name===(typeof e==`string`?e:e.name))||this._defaultModules.push(e)}removeDefaultModule(e){let t=typeof e==`string`?e:e.name;this._defaultModules=this._defaultModules.filter(e=>e.name!==t)}addShaderHook(e,t){t&&(e=Object.assign(t,{hook:e})),this._hookFunctions.push(e)}_getModuleList(e=[]){let t=Array(this._defaultModules.length+e.length),n={},r=0;for(let e=0,i=this._defaultModules.length;e<i;++e){let i=this._defaultModules[e],a=i.name;t[r++]=i,n[a]=!0}for(let i=0,a=e.length;i<a;++i){let a=e[i],o=a.name;n[o]||(t[r++]=a,n[o]=!0)}return t.length=r,L(t),t}},xn=class extends bn{shaderLanguage=`glsl`;assembleGLSLShaderPair(e){let t=this._getModuleList(e.modules),n=this._hookFunctions;return{...Wt({...e,vs:e.vs,fs:e.fs,modules:t,hookFunctions:n}),modules:t}}},Sn=class e extends bn{shaderLanguage=`wgsl`;_wgslBindingRegistry=new Map;assembleWGSLShader(t){let n=this._getModuleList(t.modules),r=this._hookFunctions,i=e.getShaderPreprocessorDefines(t,n),a=t.platformInfo.shaderLanguage===`wgsl`&&t.source?W(t.source,{defines:i}):t.source,{source:o,getUniforms:s,bindingAssignments:c}=Ut({...t,source:a,defines:i,_bindingRegistry:this._wgslBindingRegistry,modules:n,hookFunctions:r}),l=t.platformInfo.shaderLanguage===`wgsl`?W(o,{defines:i}):o;return{source:l,getUniforms:s,modules:n,bindingAssignments:c,bindingTable:Je(l,c),shaderLayout:f(l,{vertexEntryPoint:t.vertexEntryPoint,scanVertexAttributes:t.scanVertexAttributes})}}static getShaderPreprocessorDefines(t,n){return{...e.getPlatformPreprocessorDefines(t.platformInfo),...n.reduce((e,t)=>(Object.assign(e,t.defines),e),{}),...t.defines}}static getPlatformPreprocessorDefines(e){let t=e.limits||{};return{LUMA_SUPPORTS_VERTEX_STORAGE_BUFFERS:e.type===`webgpu`&&(t.maxStorageBuffersInVertexStage||0)>0,LUMA_FP32_TAN_PRECISION_WORKAROUND:e.type===`webgpu`&&e.gpu.toLowerCase()!==`nvidia`&&e.gpu.toLowerCase()!==`amd`,LUMA_FP64_INTEGER_ARITHMETIC:e.type===`webgpu`&&e.gpu.toLowerCase()===`apple`}}},Cn={name:`fp32`,source:`#ifdef LUMA_FP32_TAN_PRECISION_WORKAROUND
const FP32_TWO_PI: f32 = 6.2831854820251465;
const FP32_PI_2: f32 = 1.5707963705062866;
const FP32_PI_16: f32 = 0.1963495463132858;

const FP32_SIN_TABLE_0: f32 = 0.19509032368659973;
const FP32_SIN_TABLE_1: f32 = 0.3826834261417389;
const FP32_SIN_TABLE_2: f32 = 0.5555702447891235;
const FP32_SIN_TABLE_3: f32 = 0.7071067690849304;

const FP32_COS_TABLE_0: f32 = 0.9807852506637573;
const FP32_COS_TABLE_1: f32 = 0.9238795042037964;
const FP32_COS_TABLE_2: f32 = 0.8314695954322815;
const FP32_COS_TABLE_3: f32 = 0.7071067690849304;

const FP32_INVERSE_FACTORIAL_3: f32 = 1.666666716337204e-01;
const FP32_INVERSE_FACTORIAL_5: f32 = 8.333333767950535e-03;
const FP32_INVERSE_FACTORIAL_7: f32 = 1.9841270113829523e-04;
const FP32_INVERSE_FACTORIAL_9: f32 = 2.75573188446287533e-06;
const FP32_OVERFLOW: f32 = 3.402823466e+38;

fn sin_taylor_fp32(a: f32) -> f32 {
  if (a == 0.0) {
    return 0.0;
  }

  let x = -a * a;
  var sum = a;
  var term = a;

  term = term * x;
  sum = sum + term * FP32_INVERSE_FACTORIAL_3;
  term = term * x;
  sum = sum + term * FP32_INVERSE_FACTORIAL_5;
  term = term * x;
  sum = sum + term * FP32_INVERSE_FACTORIAL_7;
  term = term * x;
  sum = sum + term * FP32_INVERSE_FACTORIAL_9;

  return sum;
}

fn tan_taylor_fp32(a: f32) -> f32 {
  if (a == 0.0) {
    return 0.0;
  }

  let z = floor(a / FP32_TWO_PI);
  let reduced = a - FP32_TWO_PI * z;

  var quadrantValue = floor(reduced / FP32_PI_2 + 0.5);
  let quadrant = i32(quadrantValue);
  if (quadrant < -2 || quadrant > 2) {
    return FP32_OVERFLOW;
  }

  var angle = reduced - FP32_PI_2 * quadrantValue;
  quadrantValue = floor(angle / FP32_PI_16 + 0.5);
  let tableIndex = i32(quadrantValue);
  let absoluteTableIndex = abs(tableIndex);
  if (absoluteTableIndex > 4) {
    return FP32_OVERFLOW;
  }

  angle = angle - FP32_PI_16 * quadrantValue;
  let sinAngle = sin_taylor_fp32(angle);
  let cosAngle = sqrt(1.0 - sinAngle * sinAngle);

  var tableCos = 0.0;
  var tableSin = 0.0;
  if (absoluteTableIndex == 1) {
    tableCos = FP32_COS_TABLE_0;
    tableSin = FP32_SIN_TABLE_0;
  } else if (absoluteTableIndex == 2) {
    tableCos = FP32_COS_TABLE_1;
    tableSin = FP32_SIN_TABLE_1;
  } else if (absoluteTableIndex == 3) {
    tableCos = FP32_COS_TABLE_2;
    tableSin = FP32_SIN_TABLE_2;
  } else if (absoluteTableIndex == 4) {
    tableCos = FP32_COS_TABLE_3;
    tableSin = FP32_SIN_TABLE_3;
  }

  var sinReduced = sinAngle;
  var cosReduced = cosAngle;
  if (tableIndex > 0) {
    sinReduced = tableCos * sinAngle + tableSin * cosAngle;
    cosReduced = tableCos * cosAngle - tableSin * sinAngle;
  } else if (tableIndex < 0) {
    sinReduced = tableCos * sinAngle - tableSin * cosAngle;
    cosReduced = tableCos * cosAngle + tableSin * sinAngle;
  }

  var sinValue = 0.0;
  var cosValue = 0.0;
  if (quadrant == 0) {
    sinValue = sinReduced;
    cosValue = cosReduced;
  } else if (quadrant == 1) {
    sinValue = cosReduced;
    cosValue = -sinReduced;
  } else if (quadrant == -1) {
    sinValue = -cosReduced;
    cosValue = sinReduced;
  } else {
    sinValue = -sinReduced;
    cosValue = -cosReduced;
  }

  return sinValue / cosValue;
}

fn tan_fp32(a: f32) -> f32 {
  return tan_taylor_fp32(a);
}
#else
fn tan_fp32(a: f32) -> f32 {
  return tan(a);
}
#endif
`,vs:`#ifdef LUMA_FP32_TAN_PRECISION_WORKAROUND

// All these functions are for substituting tan() function from Intel GPU only
const float TWO_PI = 6.2831854820251465;
const float PI_2 = 1.5707963705062866;
const float PI_16 = 0.1963495463132858;

const float SIN_TABLE_0 = 0.19509032368659973;
const float SIN_TABLE_1 = 0.3826834261417389;
const float SIN_TABLE_2 = 0.5555702447891235;
const float SIN_TABLE_3 = 0.7071067690849304;

const float COS_TABLE_0 = 0.9807852506637573;
const float COS_TABLE_1 = 0.9238795042037964;
const float COS_TABLE_2 = 0.8314695954322815;
const float COS_TABLE_3 = 0.7071067690849304;

const float INVERSE_FACTORIAL_3 = 1.666666716337204e-01; // 1/3!
const float INVERSE_FACTORIAL_5 = 8.333333767950535e-03; // 1/5!
const float INVERSE_FACTORIAL_7 = 1.9841270113829523e-04; // 1/7!
const float INVERSE_FACTORIAL_9 = 2.75573188446287533e-06; // 1/9!

float sin_taylor_fp32(float a) {
  float r, s, t, x;

  if (a == 0.0) {
    return 0.0;
  }

  x = -a * a;
  s = a;
  r = a;

  r = r * x;
  t = r * INVERSE_FACTORIAL_3;
  s = s + t;

  r = r * x;
  t = r * INVERSE_FACTORIAL_5;
  s = s + t;

  r = r * x;
  t = r * INVERSE_FACTORIAL_7;
  s = s + t;

  r = r * x;
  t = r * INVERSE_FACTORIAL_9;
  s = s + t;

  return s;
}

void sincos_taylor_fp32(float a, out float sin_t, out float cos_t) {
  if (a == 0.0) {
    sin_t = 0.0;
    cos_t = 1.0;
  }
  sin_t = sin_taylor_fp32(a);
  cos_t = sqrt(1.0 - sin_t * sin_t);
}

float tan_taylor_fp32(float a) {
    float sin_a;
    float cos_a;

    if (a == 0.0) {
        return 0.0;
    }

    // 2pi range reduction
    float z = floor(a / TWO_PI);
    float r = a - TWO_PI * z;

    float t;
    float q = floor(r / PI_2 + 0.5);
    int j = int(q);

    if (j < -2 || j > 2) {
        return 1.0 / 0.0;
    }

    t = r - PI_2 * q;

    q = floor(t / PI_16 + 0.5);
    int k = int(q);
    int abs_k = int(abs(float(k)));

    if (abs_k > 4) {
        return 1.0 / 0.0;
    } else {
        t = t - PI_16 * q;
    }

    float u = 0.0;
    float v = 0.0;

    float sin_t, cos_t;
    float s, c;
    sincos_taylor_fp32(t, sin_t, cos_t);

    if (k == 0) {
        s = sin_t;
        c = cos_t;
    } else {
        if (abs(float(abs_k) - 1.0) < 0.5) {
            u = COS_TABLE_0;
            v = SIN_TABLE_0;
        } else if (abs(float(abs_k) - 2.0) < 0.5) {
            u = COS_TABLE_1;
            v = SIN_TABLE_1;
        } else if (abs(float(abs_k) - 3.0) < 0.5) {
            u = COS_TABLE_2;
            v = SIN_TABLE_2;
        } else if (abs(float(abs_k) - 4.0) < 0.5) {
            u = COS_TABLE_3;
            v = SIN_TABLE_3;
        }
        if (k > 0) {
            s = u * sin_t + v * cos_t;
            c = u * cos_t - v * sin_t;
        } else {
            s = u * sin_t - v * cos_t;
            c = u * cos_t + v * sin_t;
        }
    }

    if (j == 0) {
        sin_a = s;
        cos_a = c;
    } else if (j == 1) {
        sin_a = c;
        cos_a = -s;
    } else if (j == -1) {
        sin_a = -c;
        cos_a = s;
    } else {
        sin_a = -s;
        cos_a = -c;
    }
    return sin_a / cos_a;
}
#endif

float tan_fp32(float a) {
#ifdef LUMA_FP32_TAN_PRECISION_WORKAROUND
  return tan_taylor_fp32(a);
#else
  return tan(a);
#endif
}
`};export{B as a,Sn as i,xn as n,z as o,bn as r,R as s,Cn as t};