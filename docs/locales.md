# Web console locales

This document is the standing guide for **web-console translations**: how they load, how keys are used in the UI, and how to write or review them. Use it in later AI sessions and human reviews. Do not translate from another locale; **English (`en_GB`) is the source of truth**.

---

## 1. Files and loading

| Item | Location |
| --- | --- |
| Locale YAML | `web-console/public/locales/<lng>.yaml` |
| Language picker list | `web-console/src/i18n/languages.js` |
| i18n init / HTTP load | `web-console/src/i18n/i18n.js` |
| Default language | `en_GB` (`DEFAULT_LANGUAGE` in `web-console/src/Constants/Common.js`) |
| Runtime URL | `/locales/{{lng}}.yaml` (Vite public dir; packed into the server binary at release) |

The UI uses `i18next` + `react-i18next` + `i18next-http-backend`. YAML is parsed with `js-yaml`. Detection order is Redux stored language, then the browser; unknown codes fall back to `en_GB`. Hyphens are normalized to underscores (`en-GB` → `en_GB`).

Every registered language **must** have a YAML file with the **same keys** as `en_GB.yaml`. As of this writing that is **984 keys** (flat count, including nested `dialog.*`, `helper_text.*`, `error.*`, `opts.*`).

### Current languages

| Code | Language | Notes |
| --- | --- | --- |
| `en_GB` | English (UK) | Source of truth |
| `en_US` | English (US) | Spelling only (`Colour` → `Color`, `Centre` → `Center`, `Favourite` → `Favorite`) |
| `de_DE` | Deutsch | |
| `es_ES` | Español (Spain) | Formal *usted* in dialogs |
| `fr_FR` | Français | Button labels are infinitives |
| `he_IL` | Hebrew | RTL; YAML values are still plain strings |
| `hi_IN` | Hindi | Indic style (see §4) |
| `it_IT` | Italiano | Formal *Lei* in dialogs |
| `kn_IN` | Kannada | Indic style |
| `ml_IN` | Malayalam | Indic style |
| `nl_NL` | Nederlands | |
| `pl_PL` | Polski | |
| `pt_PT` | Português (Portugal) | European Portuguese, not Brazilian |
| `ro_RO` | Română | Keep ș ț ă î â |
| `ru_RU` | Русский | |
| `ta_IN` | Tamil | Indic style |
| `te_IN` | Telugu | Indic style |
| `zh_CN` | 中文 (Simplified) | PRC software terms |
| `zh_TW` | 中華民國國語 | Taiwan terms; do not mix with CN |

To add a language: copy `en_GB.yaml` → `<lng>.yaml`, translate, then append an entry to `languages.js` (`lng` must match the filename without `.yaml`).

---

## 2. How keys reach the UI

Keys are **not** always passed through `t()` at the call site. Review meaning in the screen that shows the string.

| Pattern | What happens | Typical files |
| --- | --- | --- |
| `t("key")` | Direct lookup | `Layout.jsx`, `Login.jsx`, `Actions.jsx`, `Dialog.jsx` |
| `<PageTitle title="schedules" />` | `PageTitle` calls `t(title)` | List/add pages |
| `title: "policies"` on a route | Sidebar calls `t(m.title)` | `Service/Routes.js` |
| Form item `label: "latitude"` | `Form.jsx` calls `t(item.label)` | Update pages, widget editors |
| `helperTextInvalid: "helper_text.invalid_latitude"` | Same, via `t()` | Update pages |
| `label: "opts.chart_type.label_line_chart"` | Option lists in `Constants/` | Dropdowns / radios |
| `deleteDialogTitle: "dialog.delete_title_node"` | List delete confirm | `* /ListPage.jsx` |

Nested YAML becomes dotted keys: `dialog.restore_msg`, `helper_text.invalid_id`, `opts.handler_type.desc_backup`.

`opts.*` keys with the same leaf name (`desc_none`, `label_none`, `desc_disk`) live under **different parents**. Those parents are different meanings (dampening vs logger vs backup provider). Translate in the parent context; do not assume one `desc_none` fits all.

---

## 3. Never translate

Leave these **exactly** as in English:

| Item | Keys / examples |
| --- | --- |
| Product slogan | `the_open_source_controller` → `The Open Source Controller` |
| License name | `apache_license_2` → `Apache License 2.0` |
| Binary states | `'on'` / `'off'` → `ON` / `OFF` (also `on_text`, `off_text`, `on_button`, `payload_on`, `payload_off`) |
| Booleans | `'true'` / `'false'` → `True` / `False` |
| HTTP methods | `opts.http_method.*` → `GET` `POST` `PUT` `DELETE` |
| Protocols / brands | MQTT, HTTP, Ethernet, Serial, ESPHome, Tasmota, MySensors, Alexa, … |
| Chart interpolation ids | `Basis`, `Cardinal`, `CatmullRom`, `Linear`, `StepAfter`, … |
| Light codes | `CW WW`, `RGB`, `RGB CW`, `RGB CW WW` |

Do **not** swap ON and OFF. They are status values, not prose.

---

## 4. Wording style

Goals: **easy**, **meaningful**, **professional**.

- Use the standard **written software-UI register** of that language. No dialect, slang, or chat shortcuts. No overly literary or textbook words that a console user would not say.
- Judge the string by the **screen**, not a dictionary. A preposition that is correct in isolation is wrong if the field is an email header.
- Keep UI verbs consistent inside a language (French infinitives: *Ajouter*, *Supprimer*, *Enregistrer*; German/Dutch infinitives; Indic polite written verbs).
- Plurals on list titles and multi-delete dialogs must stay plural (`Delete handlers?`, not a single handler).
- Device types: professional names, not nicknames (`Air Conditioner` / *Klimaanlage* / *空调*, not `AC` / Mixie / Plug).

### Two families

**Indic (`ta_IN`, `hi_IN`, `kn_IN`, `ml_IN`, `te_IN`)**  
Everyday verbs stay native. Named IoT resources use **native-script loanwords**, not Latin English and not calques:

| English | ta | hi | kn | ml | te |
| --- | --- | --- | --- | --- | --- |
| Gateway | கேட்வே | गेटवे | ಗೇಟ್‌ವೇ | ഗേറ്റ്‌വേ | గేట్‌వే |
| Node | நோடு | नोड | ನೋಡ್ | നോഡ് | నోడ్ |
| Source | சோர்ஸ் | सोर्स | ಸೋರ್ಸ್ | സോഴ്‌സ് | సోర్స్ |
| Field | ஃபீல்டு | फील्ड | ಫೀಲ್ಡ್ | ഫീൽഡ് | ఫీల్డ్ |
| Firmware | ஃபார்ம்வேர் | फर्मवेयर | ಫರ್ಮ್‌ವೇರ್ | ഫേംവെയർ | ఫర్మ్‌వేర్ |
| Data repository | டேட்டா ரிப்பாசிட்டரி | डेटा रिपॉजिटरी | ಡೇಟಾ ರಿಪಾಸಿಟರಿ | ഡാറ്റ റിപ്പോസിറ്ററി | డేటా రిపాజిటరీ |
| Resource | ரிசோர்ஸ் | रिसोर्स | ರಿಸೋರ್ಸ್ | റിസോഴ്‌സ് | రిసోర్స్ |
| Virtual device | விர்ச்சுவல் டிவைஸ் | वर्चुअल डिवाइस | ವರ್ಚುವಲ್ ಡಿವೈಸ್ | വെർച്വൽ ഡിവൈസ് | వర్చువల్ డివైస్ |
| Virtual assistant | விர்ச்சுவல் அசிஸ்டண்ட் | वर्चुअल असिस्टेंट | ವರ್ಚುವಲ್ ಅಸಿಸ್ಟೆಂಟ್ | വെർച്വൽ അസിസ്റ്റന്റ് | వర్చువల్ అసిస్టెంట్ |
| Payload | பேலோடு | पेलोड | ಪೇಲೋಡ್ | പേലോഡ് | పేలోడ్ |
| Forward payload | ஃபார்வர்ட் பேலோடு | फ़ॉरवर्ड पेलोड | ಫಾರ್ವರ್ಡ್ ಪೇಲೋಡ್ | ഫോർവേഡ് പേലോഡ് | ఫార్వర్డ్ పేలోడ్ |
| Quick ID | குவிக் ஐடி | क्विक आईडी | ಕ್ವಿಕ್ ಐಡಿ | ക്വിക്ക് ഐഡി | క్విక్ ఐడి |

Do not use calques such as Tamil நுழைவாயில் / முனை / அடையாளம் for those names. Do not write Latin `Gateway` inside a Tamil sentence. Keep product names and protocol abbreviations Latin (MyController, MQTT, HTTP, JSON, API).

Plurals attach without a hyphen (`கேட்வேகள்`, not `கேட்வே-கள்`). Kannada/Telugu use ZWNJ before the suffix (`ನೋಡ್‌ಗಳು`, `నోడ్‌లు`); Malayalam uses `-ുകൾ` (`നോഡുകൾ`). Hindi often keeps the same form for singular and plural loanwords (`गेटवे`). Do not reuse a clearly singular form as a list title in ta/kn/ml/te.

Other UI verbs (Save, Backup, Restore, Upload, Reboot, Reset) stay native and distinct.

**European, Hebrew, Chinese**  
Use the established professional IT term in that language. Keep a loanword when it is clearer than a coined calque (`Payload`, `Gateway`, `Handler`, `Dashboard`, `Widget`, `Firmware`, `Token` are often better than *carga útil*, *Pasarela*, *承載*, *מטען*).

Do **not** use `Dashboard` and widget `Panel` as the same word. Dashboard is the page; Panel is a widget type.

`zh_CN` and `zh_TW` must not mix vocabulary (添加 vs 新增, 字段 vs 欄位, 恢复 vs 還原, 计划 vs 排程, 登录 vs 登入).

---

## 5. Verbs that must stay distinct

These are different actions. Translating two of them to the same word is a bug.

| English | Typical failure |
| --- | --- |
| Save | Confused with Backup |
| Backup | Confused with Save |
| Restore | Incomplete or same as Backup |
| Upload | French *télécharger* is download; use *téléverser* |
| Download | Russian *Загрузить* is upload; download is *Скачать* |

`run_backup` / `run_a_backup` mean **run a backup**, not “make a copy”.

---

## 6. Meaning-critical keys (review with the UI)

Translate these only after reading the call site.

| Key | English sense | Where it shows | Trap |
| --- | --- | --- | --- |
| `from` | Email **sender** | `Components/Form/ResourcePicker/ResourcePicker.jsx` (`getEmailDataItems`) | Not the preposition “from” |
| `to` | Email **recipient** | same | Not a generic “to” |
| `subject` | Email subject | same | Not “topic” / “object” in general |
| `forward_payload` | Feature **name** | `Service/Routes.js`, `Pages/Operations/ForwardPayload/ListPage.jsx` | Not “send payload” |
| `frequency` | Schedule **how often** | `Pages/Operations/Schedule/UpdatePage.jsx` | Not radio Hz (German: *Häufigkeit*, not *Frequenz*) |
| `schedule` / `schedules` | Timed job | Operations nav + schedule pages | Must not collide with Table |
| `table_view` / `table_configuration` | Widget table | `Widgets/ControlPanel/Edit.js`, `UtilizationPanel/Table/Edit.js` | Must not collide with Schedule |
| `policies` / `policy_details` | IAM access policy | `Routes.js`, `Pages/Settings/Policy/`, user form | Not a government “doctrine” |
| `statement` / `statements` | IAM allow rule | Policy + service-account forms | Not a report |
| `hide_border` / `hide_header` | Hide widget chrome | Control / utilization widget editors | Border, not “limit” or “fort” |
| `restore` | Restore from backup | `Components/Actions/Actions.jsx`, `Dialog.jsx` | Must be a usable button label |
| `reboot` | Reboot selected nodes | Node list actions + `Dialog.jsx` | Same sense as English Reboot |
| `dialog.delete_title_*` | Delete **selected** rows | List pages | Keep plural when English is plural |
| `dialog.restore_msg` | Restore side effects | Backup page confirm | Last bullet: **you** may need to start the server **manually** |
| `dialog.node_reset_msg` | Factory reset warning | Node list | May lose access in MyController |
| `no_dashboard_message` | No dashboards **exist** | `Pages/Dashboard/Dashboard.jsx` | Not “no dashboard is currently loaded” |
| `fetching` | Loading word only | `Layout/Login.jsx` (`${t("fetching")}...`) | Do **not** put `...` in the YAML |
| `trigger_on_event` | Task runs on an event | Task list + `UpdatePage.jsx` | Not “turn the device ON” |
| `latitude` / `longitude` | Geo fields | `Pages/Settings/System/` | Settings labels, not textbook geography |
| `alive_check_interval` | Node liveness poll | ESPHome / gateway node config | Not a medical “heartbeat” |
| `keep_me_logged_in_30_days` | Remember session | `Layout/Login.jsx` | Imperative / checkbox sense |
| `documentation` | Help link | Login footer + `Layout.jsx` | Not “a document” |
| `experimental` | Badge | Sidebar (`Layout.jsx`) | Short badge text |
| `read_only` | Field not writable | Topology sidebar, forms | German software term is *Schreibgeschützt* |
| `parameters_to_handler` | Params sent to a Handler | Task / Schedule update | Grammar must fit the loanword (IT: *l'Handler*) |
| `opts.trait_type.label_transport_control` | Media play control | Virtual device traits | Not road/traffic transport |
| `opts.metric_type.label_binary` | On/off metric | Field metrics | Not “binary file format” |

`dialog.restore_msg` and `dialog.node_reset_msg` embed markup. Copy placeholders **exactly** from `en_GB`:

- `{{fileName}}`
- `<br/>`
- `<3>…</3>` wrapper
- `<0>…</0>`, `<1>…</1>`, `<2>…</2>` (and `<3>…</3>` inside restore)

---

## 7. YAML rules

- Quote keys that YAML would treat as booleans: `'on'`, `'off'`, `'true'`, `'false'`.
- Do not introduce **duplicate keys** at the same level. YAML keeps the last one and silently drops the first (this has already caused wrong meanings for `desc_disk` / `desc_none` when they were flattened).
- Apostrophes in values are fine in a plain scalar (`Parametri per l'Handler`). If a value starts with a quote or contains `: ` in a confusing way, quote the whole string.
- Keep key order close to `en_GB.yaml` so diffs stay reviewable.
- Nested maps: `dialog`, `error.console`, `helper_text`, `opts.<group>`.

---

## 8. Review checklist

Use this for AI or human review of a locale change.

**Mechanical**

1. Key count matches `en_GB` (978). No missing, no extra.
2. Placeholders on every key match `en_GB` (`{{…}}` and `<n>` tags).
3. `the_open_source_controller` is exactly `The Open Source Controller`.
4. `apache_license_2` is `Apache License 2.0`.
5. `'on'` / `'off'` are `ON` / `OFF` and not swapped.
6. Save, Backup, Restore, Upload, Download are four different words.
7. YAML parses (`yaml.Unmarshal` from `gopkg.in/yaml.v3`). Quoted boolean keys still strings.

**Meaning (spot-check the table in §6)**

8. `from` / `to` / `subject` work as email fields.
9. `forward_payload` is a feature name.
10. `frequency` is schedule cadence.
11. `schedule*` ≠ `table_*`.
12. Multi-delete titles stay plural.
13. `fetching` has no trailing `...`.
14. Restore last bullet still says the **user** may need to start the server.
15. Empty dashboard message is “none available”, not “none loaded”.
16. Indic files still use English product terms; CN/TW vocabulary is not mixed.

**Style**

17. Professional written UI language; no slang or very local spoken forms.
18. No two different English concepts collapsed into one native word.

A small Go check for items 1–2 (same `gopkg.in/yaml.v3` the server already uses). Run from the repo root after replacing `XX_YY.yaml`:

```go
package main

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

func flatten(in any, prefix string, out map[string]struct{}) {
	m, ok := in.(map[string]any)
	if !ok {
		out[prefix] = struct{}{}
		return
	}
	for k, child := range m {
		p := k
		if prefix != "" {
			p = prefix + "." + k
		}
		flatten(child, p, out)
	}
}

func loadKeys(path string) (map[string]struct{}, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw any
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	flatten(raw, "", out)
	return out, nil
}

func missing(a, b map[string]struct{}) []string {
	var keys []string
	for k := range a {
		if _, ok := b[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func main() {
	en, err := loadKeys("web-console/public/locales/en_GB.yaml")
	if err != nil {
		panic(err)
	}
	other, err := loadKeys("web-console/public/locales/XX_YY.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println("en", len(en), "other", len(other))
	fmt.Println("missing", missing(en, other))
	fmt.Println("extra", missing(other, en))
}
```

---

## 9. Adding or changing a string in code

1. Add the key to **`en_GB.yaml` first**, then `en_US.yaml` (US spelling if needed), then every other locale.
2. Prefer an existing key over a new one.
3. If the English string is a **noun** (page title, resource type), keep the translation a noun. If it is a **button**, keep it a verb in that language’s UI convention.
4. If the UI already appends punctuation (`${t("fetching")}...`, `${t("browse")}...`), do not put the same punctuation in YAML.
5. Do not build sentences by concatenating separately translated fragments unless the English source already does that.

---

## 10. Related code

| Area | Path |
| --- | --- |
| Sidebar titles | `web-console/src/Service/Routes.js` |
| Login | `web-console/src/Layout/Login.jsx` |
| Header / help / slogan | `web-console/src/Layout/Layout.jsx` |
| Action + confirm dialogs | `web-console/src/Components/Actions/Actions.jsx`, `Components/Dialog/Dialog.jsx` |
| Form label translation | `web-console/src/Components/Form/Form.jsx` |
| Option catalogs (`opts.*`) | `web-console/src/Constants/` (Gateway, Handler, Schedule, Task, VirtualDevice, Widgets) |
| Email From/To/Subject | `web-console/src/Components/Form/ResourcePicker/ResourcePicker.jsx` |
