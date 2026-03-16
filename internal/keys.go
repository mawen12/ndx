package internal

type ContextKey string

const (
	KeyApp ContextKey = "app"
)

type ComponentKey string

const (
	QueryLabelComponent      ComponentKey = "query_label"
	QueryComponent           ComponentKey = "query"
	TimeLabelComponent       ComponentKey = "time_label"
	EditBtnComponent         ComponentKey = "editBtn"
	HistogramComponent       ComponentKey = "histogram"
	TableComponent           ComponentKey = "table"
	StatusLineLeftComponent  ComponentKey = "status_line_left"
	StatusLineRightComponent ComponentKey = "status_line_right"
	StatusLineComponent      ComponentKey = "status_line"
	CmdComponent             ComponentKey = "cmd"

	EditViewTimeLabelComponent        ComponentKey = "edit_view_time_label"
	EditViewTimeComponent             ComponentKey = "edit_view_time"
	EditViewQueryLabelComponent       ComponentKey = "edit_view_query_label"
	EditViewQueryComponent            ComponentKey = "edit_view_query"
	EditViewLogStreamLabelComponent   ComponentKey = "edit_view_log_streams_label"
	EditViewLogStreamComponent        ComponentKey = "edit_view_log_streams"
	EditViewSelectQueryLabelComponent ComponentKey = "edit_view_select_query_label"
	EditViewSelectQueryComponent      ComponentKey = "edit_view_select_query"

	ConnectModalContentComponent ComponentKey = "connect_modal_content"
)

type PageKey string

const (
	KeyMainPage     PageKey = "main_page"
	KeyEditModal    PageKey = "edit_modal"
	KeyConnectModal PageKey = "connect_modal"
)
