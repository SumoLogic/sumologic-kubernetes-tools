package tracingconfig

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Migrate(input string) (string, error) {
	inputValues, err := parseValues(input)
	if err != nil {
		return "", fmt.Errorf("error parsing input yaml: %v", err)
	}

	if &inputValues.Otelcol != nil {
		outputValues, err := migrate(&inputValues)
		if err != nil {
			return "", fmt.Errorf("error migrating: %v", err)
		}

		buffer := bytes.Buffer{}
		encoder := yaml.NewEncoder(&buffer)
		encoder.SetIndent(2)
		err = encoder.Encode(outputValues)
		fmt.Sprintln(buffer.String())
		fmt.Println("WARNING! Tracing config migrated to v3, please check the output file. For more details see documentation: https://github.com/SumoLogic/sumologic-kubernetes-collection/blob/main/docs/v3-migration-doc.md#tracinginstrumentation-changes")
		return buffer.String(), err
	}

	return input, err
}

func parseValues(input string) (ValuesInput, error) {
	var outputValues ValuesInput
	err := yaml.Unmarshal([]byte(input), &outputValues)
	return outputValues, err
}

func migrate(inputValues *ValuesInput) (ValuesOutput, error) {
	outputValues := ValuesOutput{
		Rest: inputValues.Rest,
	}

	// otelcol-instrumentation migrations
	// migrate otelcol source processor to otelcol-instrumentation
	outputValues.OtelcolInstrumentation.Config.Processors.Source = inputValues.Otelcol.Config.Processors.Source

	// tracesgateway (old otelgateway) migrations
	// migrate deployment
	outputValues.TracesGateway.Deployment = inputValues.Otelgateway.Deployment
	// migrate loadbalancing exporter compression
	outputValues.TracesGateway.Config.Exporters.LoadBalancing.Protocol.Otlp.Compression = inputValues.Otelgateway.Config.Exporters.LoadBalancing.Protocol.Otlp.Compression
	// migration loadbalancing exporter num of consumers
	outputValues.TracesGateway.Config.Exporters.LoadBalancing.Protocol.Otlp.SendingQueue.NumConsumers = inputValues.Otelgateway.Config.Exporters.LoadBalancing.Protocol.Otlp.SendingQueue.NumConsumers
	// migration loadbalancing exporter queue size
	outputValues.TracesGateway.Config.Exporters.LoadBalancing.Protocol.Otlp.SendingQueue.QueueSize = inputValues.Otelgateway.Config.Exporters.LoadBalancing.Protocol.Otlp.SendingQueue.QueueSize
	// migrate batch processor
	outputValues.TracesGateway.Config.Processors.Batch = inputValues.Otelgateway.Config.Processors.Batch
	// migrate memory limiter processor
	outputValues.TracesGateway.Config.Processors.MemoryLimiter = inputValues.Otelgateway.Config.Processors.MemoryLimiter

	// tracessampler (old oltecol) migrations
	// migrate deployment
	outputValues.TracesSampler.Deployment = inputValues.Otelcol.Deployment
	// migrate cascading_filter processor
	outputValues.TracesSampler.Config.Processors.CascadingFilter = inputValues.Otelcol.Config.Processors.CascadingFilter
	// migrate batch processor
	outputValues.TracesSampler.Config.Processors.Batch = inputValues.Otelcol.Config.Processors.Batch
	// migrate memory limiter processor
	outputValues.TracesSampler.Config.Processors.MemoryLimiter = inputValues.Otelcol.Config.Processors.MemoryLimiter

	return outputValues, nil
}
