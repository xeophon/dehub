package tui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dehub/v4/internal/config"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/actionssection"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/selection"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/table"
)

func (m *Model) handleMouseClick(x, y int, regionID string) tea.Cmd {
	if m.ctx == nil {
		return nil
	}

	footerTop := common.TabsHeight + m.getBaseContentHeight()
	if y >= footerTop {
		const helpIndicatorWidth = 8
		if x >= max(0, m.ctx.ScreenWidth-helpIndicatorWidth) {
			m.footer.ShowAll = !m.footer.ShowAll
			m.syncMainContentDimensions()
		}
		return nil
	}

	if y == 0 {
		if sectionID, ok := m.tabs.SectionAtX(x); ok && m.getSectionAt(sectionID) != nil {
			m.setCurrSectionId(sectionID)
			return m.onViewedRowChanged()
		}
		return nil
	}

	if m.ctx.View == config.ActionsView {
		if as, ok := m.getCurrSection().(*actionssection.Model); ok {
			return m.handleActionsMouseClick(x, y, regionID, as)
		}
		return nil
	}

	contentTop := m.copySelectionContentY()
	if row, ok := rowIndexFromSelectionRegion(regionID); ok {
		m.setActivePane(mainPane)
		return m.selectSectionRow(m.getCurrSection(), row)
	}

	if m.sidebar.IsOpen && m.previewPaneBounds(contentTop).Contains(x, y) {
		m.setActivePane(previewPane)
		return nil
	}

	if y >= contentTop && y < contentTop+m.ctx.MainContentHeight &&
		x >= 0 && x < m.ctx.MainContentWidth {
		m.setActivePane(mainPane)
	}
	return nil
}

func (m *Model) handleActionsMouseClick(x, y int, regionID string, as *actionssection.Model) tea.Cmd {
	contentTop := m.copySelectionContentY()
	if y < contentTop {
		return nil
	}

	firstWidth, secondWidth, _ := actionsPaneWidths(m.ctx.ScreenWidth)
	switch {
	case x < firstWidth:
		as.SetFocusedPane(actionssection.PaneWorkflows)
		if row, ok := rowIndexFromSelectionRegion(regionID); ok {
			return m.selectActionsTableRow(as.Table.SetCurrItem, as.Table.SetRows, as.BuildRows, row)
		}
	case x < firstWidth+secondWidth:
		as.SetFocusedPane(actionssection.PaneRuns)
		if row, ok := rowIndexFromSelectionRegion(regionID); ok {
			return m.selectActionsTableRow(as.RunsTable.SetCurrItem, as.RunsTable.SetRows, as.BuildRunRows, row)
		}
	default:
		as.SetFocusedPane(actionssection.PaneDetails)
	}
	return m.onViewedRowChanged()
}

func (m *Model) selectActionsTableRow(setCurrItem func(int), setRows func([]table.Row), buildRows func() []table.Row, row int) tea.Cmd {
	if row < 0 {
		return nil
	}
	setCurrItem(row)
	setRows(buildRows())
	return m.onViewedRowChanged()
}

func (m *Model) selectSectionRow(currSection section.Section, row int) tea.Cmd {
	if currSection == nil || row < 0 || row >= currSection.NumRows() {
		return nil
	}
	start := currSection.CurrRow()
	for currSection.CurrRow() < row {
		if currSection.NextRow() == start {
			break
		}
	}
	for currSection.CurrRow() > row {
		if currSection.PrevRow() == start {
			break
		}
	}
	currSection.SetRows(currSection.BuildRows())
	return m.onViewedRowChanged()
}

func rowIndexFromSelectionRegion(regionID string) (int, bool) {
	rowPrefix := selection.ID("row") + ":"
	if !strings.HasPrefix(regionID, rowPrefix) {
		return 0, false
	}
	i := strings.LastIndex(regionID, ":")
	if i < 0 || i == len(regionID)-1 {
		return 0, false
	}
	row, err := strconv.Atoi(regionID[i+1:])
	return row, err == nil
}
