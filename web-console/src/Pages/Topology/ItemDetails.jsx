import {
  Button,
  Checkbox,
  DescriptionList,
  DescriptionListDescription,
  DescriptionListGroup,
  DescriptionListTerm,
  Divider,
  Split,
  SplitItem,
  TextInput,
  Title,
  TitleSizes,
} from "@patternfly/react-core"
import { TopologySideBar } from "@patternfly/react-topology"
import React from "react"
import { useTranslation } from "react-i18next"
import { useSelector } from "react-redux"
import { RouteLink } from "../../Components/Buttons/Buttons"
import { KeyValueMap, Labels } from "../../Components/DataDisplay/Label"
import { getStatus, getStatusBool } from "../../Components/Icons/Icons"
import { LastSeen } from "../../Components/Time/Time"
import { MetricTypeOptions } from "../../Constants/Metric"
import { getQuickId, ResourceType } from "../../Constants/ResourcePicker"
import { routeMap as rMap } from "../../Service/Routes"
import { api } from "../../Service/Api"
import FieldSidebarGraph from "./FieldSidebarGraph"
import { getFieldValue, getItem, getValue } from "../../Util/Util"

export const TOPOLOGY_WS_KEY = "topology_fields"

const hasEntries = (obj) => obj && typeof obj === "object" && Object.keys(obj).length > 0

const findNode = (nodes, gatewayId, nodeId) =>
  (nodes || []).find(
    (n) => String(n.gatewayId) === String(gatewayId) && String(n.nodeId) === String(nodeId)
  )

const findSource = (sources, gatewayId, nodeId, sourceId) =>
  (sources || []).find(
    (s) =>
      String(s.gatewayId) === String(gatewayId) &&
      String(s.nodeId) === String(nodeId) &&
      String(s.sourceId) === String(sourceId)
  )



const Row = ({ label, children }) => {
  const { t } = useTranslation()
  return (
    <DescriptionListGroup>
      <DescriptionListTerm>{t(label)}</DescriptionListTerm>
      <DescriptionListDescription>{children}</DescriptionListDescription>
    </DescriptionListGroup>
  )
}

const SidebarBlock = ({ title, ruled = false, children }) => {
  const { t } = useTranslation()
  return (
    <div className={`topology-sidebar-block${ruled ? " topology-sidebar-block-ruled" : ""}`}>
      <div className="topology-sidebar-block-title">{t(title)}</div>
      <div className="topology-sidebar-block-body">{children}</div>
    </div>
  )
}

const IdLink = ({ history, path, id, text }) => {
  if (text === undefined || text === null || text === "") {
    return <span>-</span>
  }
  if (!id) {
    return <span>{String(text)}</span>
  }
  return <RouteLink history={history} path={path} id={id} text={String(text)} />
}

const secondaryNameText = (id, name) => {
  if (name === undefined || name === null) {
    return ""
  }
  const text = String(name).trim()
  if (text === "" || text === String(id)) {
    return ""
  }
  return text
}

const IdWithName = ({ history, path, id, text, name }) => {
  const extra = secondaryNameText(text, name)
  return (
    <span className="topology-sidebar-id-with-name">
      <IdLink history={history} path={path} id={id} text={text} />
      {extra ? <span className="topology-sidebar-id-name">{extra}</span> : null}
    </span>
  )
}

const SEND_TIMEOUT_MS = 3000

const isFieldReadOnly = (field) => getValue(field, "labels.read_only", "false") === "true"

const ReadOnlyCheckbox = ({ field, onChanged }) => {
  const { t } = useTranslation()
  const fromField = isFieldReadOnly(field)
  const [override, setOverride] = React.useState(null)
  const [saving, setSaving] = React.useState(false)
  const seqRef = React.useRef(0)
  const checked = override === null ? fromField : override

  React.useEffect(() => {
    seqRef.current += 1
    setOverride(null)
    setSaving(false)
    if (onChanged) {
      onChanged(null)
    }
  }, [field.id, fromField])

  const onChange = (value) => {
    if (saving || !field.id) {
      return
    }
    const seq = seqRef.current + 1
    seqRef.current = seq
    setOverride(value)
    if (onChanged) {
      onChanged(value)
    }
    setSaving(true)
    api.field
      .get(field.id)
      .then((res) => {
        if (seq !== seqRef.current) {
          return null
        }
        const next = res.data || {}
        const labels = { ...(next.labels || {}) }
        if (value) {
          labels.read_only = "true"
        } else {
          delete labels.read_only
        }
        next.labels = labels
        return api.field.update(next)
      })
      .catch(() => {
        if (seq !== seqRef.current) {
          return
        }
        setOverride(null)
        if (onChanged) {
          onChanged(null)
        }
      })
      .finally(() => {
        if (seq === seqRef.current) {
          setSaving(false)
        }
      })
  }

  return (
    <Checkbox
      id={`field-read-only-${field.id}`}
      isChecked={checked}
      isDisabled={saving}
      onChange={onChange}
      aria-label={t("read_only")}
    />
  )
}

const FieldSendControl = ({ field }) => {
  const { t } = useTranslation()
  const raw = getValue(field, "current.value", "")
  const timestamp = getValue(field, "current.timestamp", "")
  const [draft, setDraft] = React.useState(raw === undefined || raw === null ? "" : String(raw))
  const [sending, setSending] = React.useState(false)
  const waitTsRef = React.useRef(null)
  const timeoutRef = React.useRef(null)

  const clearWait = () => {
    if (timeoutRef.current) {
      window.clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
    waitTsRef.current = null
  }

  const finishSend = () => {
    clearWait()
    setSending(false)
  }

  React.useEffect(() => {
    setDraft(raw === undefined || raw === null ? "" : String(raw))
    finishSend()
    return clearWait
  }, [field.id])

  React.useEffect(() => {
    if (!sending || waitTsRef.current === null) {
      return
    }
    if (String(timestamp) !== String(waitTsRef.current)) {
      finishSend()
    }
  }, [timestamp, sending])

  const onSend = () => {
    if (sending) {
      return
    }
    const quickId = getQuickId(ResourceType.Field, field)
    setSending(true)
    waitTsRef.current = timestamp
    timeoutRef.current = window.setTimeout(finishSend, SEND_TIMEOUT_MS)
    api.action.send({ resource: quickId, payload: draft }).catch(() => finishSend())
  }

  return (
    <Split hasGutter className="topology-sidebar-send">
      <SplitItem className="topology-sidebar-send-input">
        <TextInput
          aria-label={t("payload")}
          value={draft}
          isDisabled={sending}
          onChange={setDraft}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              onSend()
            }
          }}
        />
      </SplitItem>
      <SplitItem>
        <Button variant="primary" isDisabled={sending} isLoading={sending} onClick={onSend}>
          {t("send")}
        </Button>
      </SplitItem>
    </Split>
  )
}

const FieldPayload = ({ field, which }) => {
  const raw = getValue(field, `${which}.value`, "")
  const timestamp = getValue(field, `${which}.timestamp`, "")
  const value = getFieldValue(raw)
  const unit = field.unit && field.unit !== "none" ? ` ${field.unit}` : ""
  const display = value === "" ? "-" : `${value}${unit}`
  return (
    <span className="topology-sidebar-payload">
      <span>{display}</span>
      {timestamp ? (
        <span className="topology-sidebar-payload-time">
          <LastSeen date={timestamp} tooltipPosition="top" />
        </span>
      ) : null}
    </span>
  )
}

const DetailsList = ({ children }) => (
  <DescriptionList isHorizontal isCompact className="topology-sidebar-list">
    {children}
  </DescriptionList>
)

const GatewayDetails = ({ resource, history }) => (
  <>
  <DetailsList>
    <Row label="id">
      <IdLink history={history} path={rMap.resources.gateway.detail} id={resource.id} text={resource.id} />
    </Row>
    <Row label="description">{resource.description || "-"}</Row>
    <Row label="enabled">{getStatusBool(!!resource.enabled)}</Row>
    <Row label="provider">{getValue(resource, "provider.type", "") || "-"}</Row>
    <Row label="status">{getStatus(getValue(resource, "state.status", ""))}</Row>
    <Row label="since">
      <LastSeen date={getValue(resource, "state.since", "")} tooltipPosition="top" />
    </Row>
    <Row label="message">{getValue(resource, "state.message", "") || "-"}</Row>
  </DetailsList>
    {hasEntries(resource.labels) ? (
      <SidebarBlock title="labels" ruled>
        <Labels data={resource.labels} />
      </SidebarBlock>
    ) : null}
  </>
)

const NodeDetails = ({ resource, history }) => (
  <>
  <DetailsList>
    <Row label="gateway_id">
      <IdLink
        history={history}
        path={rMap.resources.gateway.detail}
        id={resource.gatewayId}
        text={resource.gatewayId}
      />
    </Row>
    <Row label="node_id">
      <IdLink history={history} path={rMap.resources.node.detail} id={resource.id} text={resource.nodeId} />
    </Row>
    <Row label="name">
      <IdLink history={history} path={rMap.resources.node.detail} id={resource.id} text={resource.name} />
    </Row>
      <Row label="version">{getValue(resource, "labels.version", "") || "-"}</Row>
      <Row label="library_version">{getValue(resource, "labels.library_version", "") || "-"}</Row>
      <Row label="battery">{getValue(resource, "others.battery_level", "") || "-"}</Row>
      <Row label="status">{getStatus(getValue(resource, "state.status", ""))}</Row>
      <Row label="last_seen">
        <LastSeen date={resource.lastSeen} tooltipPosition="top" />
      </Row>
  </DetailsList>
      {hasEntries(resource.labels) ? (
        <SidebarBlock title="labels" ruled>
          <Labels data={resource.labels} />
        </SidebarBlock>
      ) : null}
      {hasEntries(resource.others) ? (
        <SidebarBlock title="others">
          <KeyValueMap data={resource.others} />
        </SidebarBlock>
      ) : null}
  </>
)

const SourceDetails = ({ resource, history, nodes }) => {
  const node = findNode(nodes, resource.gatewayId, resource.nodeId)
  return (
    <>
    <DetailsList>
      <Row label="gateway_id">
        <IdLink
          history={history}
          path={rMap.resources.gateway.detail}
          id={resource.gatewayId}
          text={resource.gatewayId}
        />
      </Row>
      <Row label="node_id">
        <IdWithName
          history={history}
          path={rMap.resources.node.detail}
          id={node && node.id}
          text={resource.nodeId}
          name={node && node.name}
        />
      </Row>
      <Row label="source_id">
        <IdLink
          history={history}
          path={rMap.resources.source.detail}
          id={resource.id}
          text={resource.sourceId}
        />
      </Row>
      <Row label="name">
        <IdLink history={history} path={rMap.resources.source.detail} id={resource.id} text={resource.name} />
      </Row>
        <Row label="last_seen">
          <LastSeen date={resource.lastSeen} tooltipPosition="top" />
        </Row>
    </DetailsList>
        {hasEntries(resource.labels) ? (
          <SidebarBlock title="labels" ruled>
            <Labels data={resource.labels} />
          </SidebarBlock>
        ) : null}
        {hasEntries(resource.others) ? (
          <SidebarBlock title="others">
            <KeyValueMap data={resource.others} />
          </SidebarBlock>
        ) : null}
    </>
  )
}

const FieldDetails = ({ resource, history, nodes, sources }) => {
  const { t } = useTranslation()
  const node = findNode(nodes, resource.gatewayId, resource.nodeId)
  const source = findSource(sources, resource.gatewayId, resource.nodeId, resource.sourceId)
  const metric = getItem(resource.metricType, MetricTypeOptions, -1)
  const [readOnlyOverride, setReadOnlyOverride] = React.useState(null)
  React.useEffect(() => {
    setReadOnlyOverride(null)
  }, [resource.id])
  const writable = !(readOnlyOverride === null ? isFieldReadOnly(resource) : readOnlyOverride)
  return (
    <>
    <DetailsList>
      <Row label="gateway_id">
        <IdLink
          history={history}
          path={rMap.resources.gateway.detail}
          id={resource.gatewayId}
          text={resource.gatewayId}
        />
      </Row>
      <Row label="node_id">
        <IdWithName
          history={history}
          path={rMap.resources.node.detail}
          id={node && node.id}
          text={resource.nodeId}
          name={node && node.name}
        />
      </Row>
      <Row label="source_id">
        <IdWithName
          history={history}
          path={rMap.resources.source.detail}
          id={source && source.id}
          text={resource.sourceId}
          name={source && source.name}
        />
      </Row>
      <Row label="field_id">
        <IdLink
          history={history}
          path={rMap.resources.field.detail}
          id={resource.id}
          text={resource.fieldId}
        />
      </Row>
      <Row label="name">
        <IdLink history={history} path={rMap.resources.field.detail} id={resource.id} text={resource.name} />
      </Row>
      <Row label="metric_type">{metric ? t(metric.label) : resource.metricType || "-"}</Row>
      <Row label="last_seen">
        <LastSeen date={resource.lastSeen} tooltipPosition="top" />
      </Row>
      <Row label="no_change_since">
        <LastSeen date={resource.noChangeSince} tooltipPosition="top" />
      </Row>
      <Row label="read_only">
        <ReadOnlyCheckbox field={resource} onChanged={setReadOnlyOverride} />
      </Row>
    </DetailsList>
      <Divider className="topology-sidebar-rule" />
      <DetailsList>
        <Row label="previous_value">
          <FieldPayload field={resource} which="previous" />
        </Row>
        <Row label="value">
          <FieldPayload field={resource} which="current" />
        </Row>
      </DetailsList>
      {writable ? (
        <div className="topology-sidebar-send-row">
          <FieldSendControl field={resource} />
        </div>
      ) : null}
      <FieldSidebarGraph field={resource} />
      {hasEntries(resource.labels) ? (
        <SidebarBlock title="labels" ruled>
          <Labels data={resource.labels} />
        </SidebarBlock>
      ) : null}
      {hasEntries(resource.others) ? (
        <SidebarBlock title="others">
          <KeyValueMap data={resource.others} />
        </SidebarBlock>
      ) : null}
    </>
  )
}

const sidebarTitle = (kind, resource) => {
  if (!resource) {
    return ""
  }
  if (kind === "gateway") {
    return resource.id
  }
  if (kind === "node") {
    return resource.nodeId
  }
  if (kind === "source") {
    return resource.sourceId
  }
  if (kind === "field") {
    return resource.fieldId
  }
  return resource.id || ""
}

const ItemDetails = ({ show, onClose, kind, resource, history, nodes, sources }) => {
  const { t } = useTranslation()
  const quickId = kind === "field" && resource ? getQuickId(ResourceType.Field, resource) : ""
  const liveField = useSelector((state) => {
    if (!quickId) {
      return undefined
    }
    const bucket = state.entities.websocket.data[TOPOLOGY_WS_KEY]
    return bucket ? bucket[quickId] : undefined
  })
  const data = kind === "field" && liveField ? liveField : resource
  const header = (
    <div className="topology-sidebar-header">
      <Title headingLevel="h2" size={TitleSizes.lg}>
        {t(kind || "details")}
      </Title>
      <div className="topology-sidebar-subtitle">{sidebarTitle(kind, data)}</div>
      <Divider className="topology-sidebar-title-line" />
    </div>
  )

  let body = null
  if (data && kind === "gateway") {
    body = <GatewayDetails resource={data} history={history} />
  } else if (data && kind === "node") {
    body = <NodeDetails resource={data} history={history} />
  } else if (data && kind === "source") {
    body = <SourceDetails resource={data} history={history} nodes={nodes} />
  } else if (data && kind === "field") {
    body = <FieldDetails resource={data} history={history} nodes={nodes} sources={sources} />
  }

  return (
    <TopologySideBar className="topology-sidebar" show={show} onClose={onClose} header={header}>
      {body}
    </TopologySideBar>
  )
}

export default ItemDetails
