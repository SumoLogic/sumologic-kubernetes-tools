package events

func migrateEvents(
	commonV2 CommonEventsV2,
	fluentDV2 FluentDEventsV2,
	fluentDPersistenceV2 FluentDPersistenceV2,
	otelV2 OteleventsV2,
) (commonV3 CommonEventsV3, fluentDV3 FluentDEventsV3, otelV3 OteleventsV3) {
	// we're getting rid of fluentD config completely
	fluentDV3 = FluentDEventsV3{}

	// port the otel config directly, but without persistence config
	otelV3 = OteleventsV3{
		Rest: otelV2.Rest,
	}

	// take care of whether events are enabled
	var enabled bool
	if commonV2.Enabled != nil {
		enabled = *commonV2.Enabled
	} else if fluentDV2.Enabled != nil {
		enabled = *fluentDV2.Enabled
	} else {
		enabled = true
	}

	if enabled {
		commonV3.Enabled = nil // this is the default
	} else {
		commonV3.Enabled = &enabled
	}

	// the default provider is now otelcol
	if commonV2.Provider != nil && *commonV2.Provider == "otelcol" {
		commonV3.Provider = nil
	} else { // if the user specifically chose FluentD, keep it
		commonV3.Provider = commonV2.Provider
	}

	// migrate over sourceName and sourceCategory configuration
	if fluentDV2.SourceName != nil {
		commonV3.SourceName = fluentDV2.SourceName
	}
	if fluentDV2.SourceCategory != nil {
		commonV3.SourceCategory = fluentDV2.SourceCategory
	}

	// migrate over persistence configuration
	// use the right persistence config depending on provider
	otPersistence := otelV2.Persistence
	commonPersistence := &commonV3.Persistence
	if commonV3.Provider != nil && *commonV3.Provider == "fluentd" {
		if fluentDPersistenceV2.Enabled != nil {
			commonPersistence.Enabled = fluentDPersistenceV2.Enabled
		}
		if fluentDPersistenceV2.Size != nil {
			commonPersistence.Size = fluentDPersistenceV2.Size
		}
		if fluentDPersistenceV2.AccessMode != nil {
			commonPersistence.PersistentVolume.AccessMode = fluentDPersistenceV2.AccessMode
		}
		if fluentDPersistenceV2.StorageClass != nil {
			commonPersistence.PersistentVolume.StorageClass = fluentDPersistenceV2.StorageClass
		}
	} else {
		if otPersistence.Enabled != nil {
			commonPersistence.Enabled = otPersistence.Enabled
		}
		if otPersistence.Size != nil {
			commonPersistence.Size = otPersistence.Size
		}
		if otPersistence.AccessMode != nil {
			commonPersistence.PersistentVolume.AccessMode = otPersistence.AccessMode
		}
		if len(otPersistence.PvcLabels) != 0 {
			commonPersistence.PersistentVolume.PvcLabels = otPersistence.PvcLabels
		}
	}
	return commonV3, fluentDV3, otelV3
}

func migrateEventsFull(valuesV3 *ValuesV3, valuesV2 *ValuesV2) error {
	commonV3, fluentDV3, otelV3 := migrateEvents(valuesV2.Sumologic.Events, valuesV2.Fluentd.Events, valuesV2.Fluentd.Persistence, valuesV2.Otelevents)
	valuesV3.Sumologic.Events = commonV3
	valuesV3.Fluentd.Events = fluentDV3
	valuesV3.Otelevents = otelV3
	return nil
}
