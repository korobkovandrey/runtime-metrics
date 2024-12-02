package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func TestNewGaugeSource(t *testing.T) {
	s := NewGaugeSource()
	assert.IsType(t, s, &Source{})
	assert.Equal(t, s.Len(), len(runtimeMetricNames)+2)
	for _, m := range runtimeMetricNames {
		assert.Contains(t, s.gaugeData, m)
	}
	assert.Contains(t, s.gaugeData, randomValueName)
}

func TestSource_Collect(t *testing.T) {
	s := NewGaugeSource()
	assert.Equal(t, s.collectCount.value, uint64(0))
	assert.Empty(t, s.gaugeData[randomValueName].value)
	assert.NoError(t, s.Collect())
	assert.Equal(t, s.collectCount.value, uint64(1))
	assert.NotEmpty(t, s.gaugeData[randomValueName].value)
}

func TestSource_GetDataForSendAndSetDataSent(t *testing.T) {
	s := NewGaugeSource()
	require.NoError(t, s.Collect())
	result := s.GetDataForSend(time.Second, time.Second)
	assert.Len(t, result, len(runtimeMetricNames)+2)
	dataIndex := map[string]string{}
	sentData := make([]DataSent, 0, len(result)+1)
	for _, m := range result {
		dataIndex[m.T+`_`+m.Name] = m.Value
		sentData = append(sentData, DataSent{Sent: time.Now(), T: m.T, Name: m.Name})
	}
	s.SetDataSent(sentData)
	for _, m := range runtimeMetricNames {
		assert.Contains(t, dataIndex, `gauge_`+m)
	}
	assert.Contains(t, dataIndex, `gauge_`+randomValueName)
	assert.Contains(t, dataIndex, `counter_`+collectCountName)
	assert.NotEqual(t, dataIndex[`gauge_`+randomValueName], `0`)
	assert.Equal(t, dataIndex[`counter_`+collectCountName], `1`)

	result = s.GetDataForSend(time.Second, time.Second)
	assert.Len(t, result, 0)
	time.Sleep(500 * time.Millisecond)
	result = s.GetDataForSend(time.Second, time.Second)
	assert.Len(t, result, 0)
	time.Sleep(500 * time.Millisecond)
	result = s.GetDataForSend(time.Second, time.Second)
	assert.Len(t, result, len(runtimeMetricNames)+2)
}

func TestSource_Len(t *testing.T) {
	s := NewGaugeSource()
	assert.Equal(t, s.Len(), len(runtimeMetricNames)+2)
}

func TestSource_addCollectCountSent(t *testing.T) {
	s := NewGaugeSource()
	assert.Equal(t, s.collectCount.sentValue, uint64(0))
	s.addCollectCountSent(`10`)
	assert.Equal(t, s.collectCount.sentValue, uint64(10))
	s.addCollectCountSent(`20`)
	assert.Equal(t, s.collectCount.sentValue, uint64(30))
}

func TestSource_getDiffCollectCount(t *testing.T) {
	s := NewGaugeSource()
	assert.Equal(t, s.getDiffCollectCount(), 0)
	s.collectCount.value = 10
	s.addCollectCountSent(`10`)
	assert.Equal(t, s.getDiffCollectCount(), 0)
	s.collectCount.value = 20
	assert.Equal(t, s.getDiffCollectCount(), 10)
}
