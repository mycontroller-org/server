import { AddCircleOIcon, CheckIcon, CloseIcon, EditIcon } from "@patternfly/react-icons"
import React from "react"
import { Responsive, WidthProvider } from "react-grid-layout"
import { withTranslation } from "react-i18next"
import { connect } from "react-redux"
import Actions from "../../Components/Actions/Actions"
import { LinkButton } from "../../Components/Buttons/Buttons"
import DeleteDialog from "../../Components/Dialog/Dialog"
import EmptyState from "../../Components/EmptyState/EmptyState"
import Loading from "../../Components/Loading/Loading"
import PageContent from "../../Components/PageContent/PageContent"
import PageTitle from "../../Components/PageTitle/PageTitle"
import Selector from "../../Components/Selector/Seletor"
import EditWidget from "../../Components/Widgets/EditWidget"
import LoadWidgets from "../../Components/Widgets/LoadWidgets"
import { GridBreakPoints, GridColumns } from "../../Constants/Dashboard"
import { WidgetType } from "../../Constants/Widgets/Widgets"
import { api } from "../../Service/Api"
import { updateSelectionId } from "../../store/entities/dashboard"
import { cloneDeep, getRandomId, isEqual } from "../../Util/Util"
import "./Dashboard.scss"
import EditSettings from "./EditSettings"

const ResponsiveGridLayout = WidthProvider(Responsive)

class Dashboard extends React.Component {
  state = {
    loading: true,
    editEnabled: false,
    dashboards: [],
    dashboardOnEdit: {},
    showDeleteDialog: false,
    showEditWidget: false,
    showEditSettings: false,
    targetWidget: {},
  }

  reloadDashboards = () => {
    api.dashboard
      .list({})
      .then((res) => {
        const dashboards = res.data.data
        let selectionId = this.props.selectionId

        // verify locally stored id available
        if (selectionId !== "") {
          let found = false
          for (let index = 0; index < dashboards.length; index++) {
            if (selectionId === dashboards[index].id) {
              found = true
              break
            }
          }
          if (!found) {
            selectionId = ""
          }
        }

        // if no id set set an id
        if (selectionId === "") {
          selectionId = dashboards.length > 0 ? dashboards[0].id : ""
          this.props.updateSelectionId(selectionId)
        }
        this.setState({
          dashboards: res.data.data,
          loading: false,
          dashboardOnEdit: {},
          showDeleteDialog: false,
          showEditWidget: false,
          showEditSettings: false,
          editEnabled: false,
        })
      })
      .catch((_e) => {
        this.setState({ loading: false })
      })
  }

  componentDidMount() {
    this.reloadDashboards()
  }

  shouldComponentUpdate(nextProps, nextState) {
    return !isEqual(this.state, nextState) || !isEqual(this.props, nextProps)
  }

  onEditClick = () => {
    this.setState((prevState) => {
      const { dashboards } = prevState
      const { selectionId } = this.props
      const dashboard = getItemById(dashboards, selectionId)
      return { editEnabled: true, dashboardOnEdit: dashboard }
    })
  }

  onSaveClick = () => {
    this.setState((prevState) => {
      api.dashboard.update(prevState.dashboardOnEdit).then((_res) => {
        this.setState({ editEnabled: false }, () => {
          this.reloadDashboards()
        })
      })
      return {}
    })
  }

  onLayoutChange = (layouts) => {
    // console.log(JSON.stringify(layouts, " ", 2))
    if (!layouts || layouts.length === 0) {
      return
    }
    this.setState((prevState) => {
      const { dashboardOnEdit } = prevState
      const widgets = dashboardOnEdit.widgets
      if (!widgets) {
        return {}
      }
      for (let index = 0; index < widgets.length; index++) {
        const widget = widgets[index]
        const layout = getItemById(layouts, widget.id, "i")
        if (widget.id === layout.i) {
          widgets[index].layout = {
            ...widget.layout,
            w: layout.w,
            h: layout.h,
            x: layout.x,
            y: layout.y,
          }
        }
      }
      return { dashboardOnEdit }
    })
  }

  onCancelClick = () => {
    this.reloadDashboards()
  }

  onNewDashboardClick = () => {
    this.setState((prevState) => {
      const { dashboards } = prevState
      const dashboard = getNewDashboard(this.props.t)
      dashboards.push(dashboard)
      this.props.updateSelectionId(dashboard.id)
      return {
        dashboards: dashboards,
        dashboardOnEdit: dashboard,
        editEnabled: true,
      }
    })
  }

  onDashboardChange = (item) => {
    this.props.updateSelectionId(item.id)
  }

  onDashboardBookmarkAction = (item) => {
    //console.log(item)
    this.setState((prevState) => {
      const { dashboards } = prevState
      const clonedDashboards = cloneDeep(dashboards)
      for (let index = 0; index < clonedDashboards.length; index++) {
        const dashboard = clonedDashboards[index]
        if (item.id === dashboard.id) {
          clonedDashboards[index].favorite = !dashboard.favorite
          // update to server
          api.dashboard.update(clonedDashboards[index]).then((_res) => {
            // no action required on response
          })
          break
        }
      }
      return { dashboards: clonedDashboards }
    })
  }

  onDashboardDeleteClick = (dashboardId) => {
    console.log(dashboardId)
    api.dashboard
      .delete([dashboardId])
      .then((_res) => {
        this.reloadDashboards()
      })
      .catch((_e) => {
        // error
      })
  }

  onDeleteClick = () => {
    this.setState({ showDeleteDialog: true })
  }

  onDeleteCancel = () => {
    this.setState({ showDeleteDialog: false })
  }

  onEditSettingsClick = () => {
    this.setState({ showEditSettings: true })
  }

  onEditSettingsHide = () => {
    this.setState({ showEditSettings: false })
  }

  onSettingsConfigChange = (dashboard) => {
    this.setState({ dashboardOnEdit: dashboard })
  }

  onSettingsConfigSave = (dashboard) => {
    this.setState({ showEditSettings: false, dashboardOnEdit: dashboard })
  }

  onWidgetEdit = (config) => {
    this.setState({ showEditWidget: true, targetWidget: config })
  }

  onWidgetDelete = (widgetId) => {
    this.setState((prevState) => {
      const { dashboardOnEdit } = prevState
      const clonedDashboard = cloneDeep(dashboardOnEdit)
      const { widgets } = clonedDashboard

      const newWidgets = widgets.filter((w) => {
        return w.id !== widgetId
      })
      clonedDashboard.widgets = newWidgets
      return { dashboardOnEdit: clonedDashboard }
    })
  }

  onAddWidgetClick = () => {
    this.setState((prevState) => {
      const { dashboardOnEdit } = prevState
      const clonedDashboard = cloneDeep(dashboardOnEdit)
      const newWidget = getNewWidget(this.props.t)
      clonedDashboard.widgets.push(newWidget)
      return { dashboardOnEdit: clonedDashboard }
    })
  }

  onWidgetEditHide = () => {
    this.setState({ showEditWidget: false })
  }

  onWidgetConfigChange = (widgetConfig) => {
    this.setState({ targetWidget: widgetConfig })
  }

  onWidgetConfigSave = (widgetConfig) => {
    this.setState((prevState) => {
      const { dashboardOnEdit } = prevState
      if (dashboardOnEdit.widgets) {
        for (let index = 0; index < dashboardOnEdit.widgets.length; index++) {
          const widget = dashboardOnEdit.widgets[index]
          if (widget.id === widgetConfig.id) {
            dashboardOnEdit.widgets[index] = widgetConfig
          }
        }
      }
      return { showEditWidget: false, targetWidget: widgetConfig, dashboardOnEdit: dashboardOnEdit }
    })
  }

  render() {
    const {
      loading,
      dashboards,
      editEnabled,
      showEditWidget,
      targetWidget,
      showEditSettings,
      dashboardOnEdit,
      showDeleteDialog,
    } = this.state

    const { selectionId, t } = this.props

    const pageContent = []
    let title = ""
    const dashboardItems = []

    if (loading) {
      pageContent.push(<Loading key="loading" />)
    } else if (dashboards.length === 0) {
      // pageContent.push(<span>No dashboards found. Create new</span>)
      pageContent.push(
        <EmptyState
          key="empty-state"
          title={t("no_dashboard_title")}
          message={t("no_dashboard_message")}
          primaryAction={{ text: t("create_new"), onClick: this.onNewDashboardClick }}
        />
      )
    } else {
      let dashboard = selectionId === "" ? dashboards[0] : getItemById(dashboards, selectionId)

      if (editEnabled) {
        dashboard = dashboardOnEdit
      }

      title = dashboard.title

      // update dashboards
      dashboards.forEach((d) => {
        dashboardItems.push({
          id: d.id,
          text: d.title,
          description: d.description,
          favorite: d.favorite,
          disabled: d.disabled,
        })
      })

      const layouts = []
      dashboard.widgets.forEach((widget) => {
        layouts.push({
          i: widget.id,
          x: widget.layout.x,
          y: widget.layout.y,
          w: widget.layout.w,
          h: widget.layout.h,
          minW: 5, // minimum width
          minH: 5, // minimum height
          static: widget.static,
          isDraggable: editEnabled && !widget.static,
          isResizable: editEnabled && !widget.static,
        })
      })

      const widgetsLoaded = LoadWidgets(
        dashboard.widgets,
        editEnabled,
        this.onWidgetEdit,
        this.onWidgetDelete,
        this.props.history
      )

      pageContent.push(
        <ResponsiveGridLayout
          key="dashboardLayout"
          className="layout"
          layouts={{ lg: layouts }}
          breakpoints={GridBreakPoints}
          cols={GridColumns}
          rowHeight={1}
          autoSize={true}
          margin={[3, 3]}
          containerPadding={[0, 0]}
          onLayoutChange={this.onLayoutChange}
          preventCollision={false}
          resizeHandles={["se"]}
        >
          {widgetsLoaded}
        </ResponsiveGridLayout>
      )
    }

    const actions = []

    if (editEnabled) {
      actions.push(
        <LinkButton
          key="btn-settings"
          text={t("settings")}
          icon={EditIcon}
          onClick={this.onEditSettingsClick}
        />,
        <LinkButton
          key="btn-widgets"
          text={t("add_widget")}
          icon={AddCircleOIcon}
          onClick={this.onAddWidgetClick}
        />,
        <LinkButton key="btn-save" text={t("save")} icon={CheckIcon} onClick={this.onSaveClick} />,
        <LinkButton key="btn-cancel" text={t("cancel")} icon={CloseIcon} onClick={this.onCancelClick} />
      )
    } else {
      const disableButtons = dashboards.length === 0
      const actionItems = [
        { type: "new", onClick: this.onNewDashboardClick },
        { type: "edit", onClick: this.onEditClick, disabled: disableButtons },
        { type: "delete", onClick: this.onDeleteClick, disabled: disableButtons },
      ]
      actions.push(<Actions key="actions" items={actionItems} isDisabled={loading} />)
    }

    return (
      <React.Fragment>
        <DeleteDialog
          key="delete-confirmation"
          onCloseFn={this.onDeleteCancel}
          onOkFn={() => this.onDashboardDeleteClick(selectionId)}
          dialogTitle="dialog.delete_title_dashboard"
          show={showDeleteDialog}
        />
        <EditSettings
          key="edit-title"
          showEditSettings={showEditSettings}
          dashboard={dashboardOnEdit}
          onCancel={this.onEditSettingsHide}
          onChange={this.onSettingsConfigChange}
          onSave={this.onSettingsConfigSave}
        />
        <EditWidget
          key="edit-widget"
          showEditWidget={showEditWidget}
          widgetConfig={targetWidget}
          onCancel={this.onWidgetEditHide}
          onChange={this.onWidgetConfigChange}
          onSave={this.onWidgetConfigSave}
        />
        <div className="dashboard">
          <PageTitle
            key="page-title"
            title={
              <div style={{ marginBottom: "3px" }}>
                <Selector
                  //prefix="Dashboard"
                  disabled={editEnabled || loading || dashboardItems.length === 0}
                  items={dashboardItems}
                  selection={{ text: title }}
                  onChange={this.onDashboardChange}
                  onBookmarkAction={this.onDashboardBookmarkAction}
                />
              </div>
            }
            actions={actions}
          />
          <PageContent key="page-content">{pageContent}</PageContent>
        </div>
      </React.Fragment>
    )
  }
}

const mapStateToProps = (state) => ({
  selectionId: state.entities.dashboard.selectionId,
})

const mapDispatchToProps = (dispatch) => ({
  updateSelectionId: (data) => dispatch(updateSelectionId(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(withTranslation()(Dashboard))

// helper functions

const getItemById = (items, value, key = "id") => {
  for (let index = 0; index < items.length; index++) {
    if (items[index][key] === value) {
      return items[index]
    }
  }
  return {}
}

// creates new default dashboard
const getNewDashboard = (t) => {
  const randomId = getRandomId(5)
  const widget = getNewWidget(t)
  const dashboard = {
    id: randomId,
    type: "desktop",
    title: "new-dashboard-" + randomId,
    bookmarked: false,
    labels: {},
    widgets: [widget],
  }

  return dashboard
}

const getNewWidget = (t = () => {}) => {
  const newWidget = {
    id: getRandomId(),
    title: t("panel_title"),
    showTitle: true,
    type: WidgetType.EmptyPanel,
    static: false,
    layout: {
      w: 20,
      h: 30,
      x: 0,
      y: Infinity,
    },
    config: {},
  }
  return newWidget
}
