package main

type ValuesV2 struct {
	NameOverride     *string `yaml:"nameOverride,omitempty"`
	FullnameOverride *string `yaml:"fullnameOverride,omitempty"`
	Sumologic        struct {
		SetupEnabled   *bool   `yaml:"setupEnabled,omitempty"`
		CleanupEnabled *bool   `yaml:"cleanupEnabled,omitempty"`
		Endpoint       *string `yaml:"endpoint,omitempty"`
		HTTPProxy      *string `yaml:"httpProxy,omitempty"`
		HTTPSProxy     *string `yaml:"httpsProxy,omitempty"`
		NoProxy        *string `yaml:"noProxy,omitempty"`
		ClusterName    *string `yaml:"clusterName,omitempty"`
		Cluster        struct {
			Host                 *string `yaml:"host,omitempty"`
			ClusterCaCertificate *string `yaml:"cluster_ca_certificate,omitempty"`
			Token                *string `yaml:"token,omitempty"`
			LoadConfigFile       *bool   `yaml:"load_config_file,omitempty"`
		} `yaml:"cluster,omitempty"`
		CollectionMonitoring *bool    `yaml:"collectionMonitoring,omitempty"`
		PodLabels            struct{} `yaml:"podLabels,omitempty"`
		PodAnnotations       struct{} `yaml:"podAnnotations,omitempty"`
		ServiceAccount       struct {
			Annotations struct{} `yaml:"annotations,omitempty"`
		} `yaml:"serviceAccount,omitempty"`
		Scc struct {
			Create *bool `yaml:"create,omitempty"`
		} `yaml:"scc,omitempty"`
		Setup struct {
			Job struct {
				Image struct {
					Repository *string `yaml:"repository,omitempty"`
					Tag        *string `yaml:"tag,omitempty"`
					PullPolicy *string `yaml:"pullPolicy,omitempty"`
				} `yaml:"image,omitempty"`
				Resources struct {
					Limits struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				NodeSelector   struct{} `yaml:"nodeSelector,omitempty"`
				PodLabels      struct{} `yaml:"podLabels,omitempty"`
				PodAnnotations struct{} `yaml:"podAnnotations,omitempty"`
			} `yaml:"job,omitempty"`
		} `yaml:"setup,omitempty"`
		Collector struct {
			Fields  struct{} `yaml:"fields,omitempty"`
			Sources struct {
				Metrics struct {
					Default struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"default,omitempty"`
					Apiserver struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"apiserver,omitempty"`
					Controller struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"controller,omitempty"`
					Scheduler struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"scheduler,omitempty"`
					State struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"state,omitempty"`
					Kubelet struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"kubelet,omitempty"`
					Node struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"node,omitempty"`
					ControlPlane struct {
						Name *string `yaml:"name,omitempty"`
					} `yaml:"control-plane,omitempty"`
				} `yaml:"metrics,omitempty"`
				Logs struct {
					Default struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
					} `yaml:"default,omitempty"`
				} `yaml:"logs,omitempty"`
				Events struct {
					Default struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
						Category   *bool   `yaml:"category,omitempty"`
					} `yaml:"default,omitempty"`
				} `yaml:"events,omitempty"`
				Traces struct {
					Default struct {
						Name       *string `yaml:"name,omitempty"`
						ConfigName *string `yaml:"config-name,omitempty"`
						Properties struct {
							ContentType *string `yaml:"content_type,omitempty"`
						} `yaml:"properties,omitempty"`
					} `yaml:"default,omitempty"`
				} `yaml:"traces,omitempty"`
			} `yaml:"sources,omitempty"`
		} `yaml:"collector,omitempty"`
		Logs struct {
			Enabled  *bool `yaml:"enabled,omitempty"`
			Metadata struct {
				Provider *string `yaml:"provider,omitempty"`
			} `yaml:"metadata,omitempty"`
			Fields []*string `yaml:"fields,omitempty"`
		} `yaml:"logs,omitempty"`
		Metrics struct {
			Enabled  *bool `yaml:"enabled,omitempty"`
			Metadata struct {
				Provider *string `yaml:"provider,omitempty"`
			} `yaml:"metadata,omitempty"`
			RemoteWriteProxy struct {
				Enabled *bool `yaml:"enabled,omitempty"`
				Config  struct {
					ClientBodyBufferSize *string `yaml:"clientBodyBufferSize,omitempty"`
					WorkerCountAutotune  *bool   `yaml:"workerCountAutotune,omitempty"`
				} `yaml:"config,omitempty"`
				ReplicaCount *int `yaml:"replicaCount,omitempty"`
				Image        struct {
					Repository *string `yaml:"repository,omitempty"`
					Tag        *string `yaml:"tag,omitempty"`
					PullPolicy *string `yaml:"pullPolicy,omitempty"`
				} `yaml:"image,omitempty"`
				Resources struct {
					Limits struct {
						CPU    *string `yaml:"cpu,omitempty"`
						Memory *string `yaml:"memory,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						CPU    *string `yaml:"cpu,omitempty"`
						Memory *string `yaml:"memory,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				LivenessProbe struct {
					InitialDelaySeconds *int `yaml:"initialDelaySeconds,omitempty"`
					PeriodSeconds       *int `yaml:"periodSeconds,omitempty"`
					TimeoutSeconds      *int `yaml:"timeoutSeconds,omitempty"`
					SuccessThreshold    *int `yaml:"successThreshold,omitempty"`
					FailureThreshold    *int `yaml:"failureThreshold,omitempty"`
				} `yaml:"livenessProbe,omitempty"`
				ReadinessProbe struct {
					InitialDelaySeconds *int `yaml:"initialDelaySeconds,omitempty"`
					PeriodSeconds       *int `yaml:"periodSeconds,omitempty"`
					TimeoutSeconds      *int `yaml:"timeoutSeconds,omitempty"`
					SuccessThreshold    *int `yaml:"successThreshold,omitempty"`
					FailureThreshold    *int `yaml:"failureThreshold,omitempty"`
				} `yaml:"readinessProbe,omitempty"`
			} `yaml:"remoteWriteProxy,omitempty"`
		} `yaml:"metrics,omitempty"`
		Traces struct {
			Enabled         *bool `yaml:"enabled,omitempty"`
			SpansPerRequest *int  `yaml:"spans_per_request,omitempty"`
		} `yaml:"traces,omitempty"`
	} `yaml:"sumologic,omitempty"`
	Fluentd struct {
		Image struct {
			Repository *string `yaml:"repository,omitempty"`
			Tag        *string `yaml:"tag,omitempty"`
			PullPolicy *string `yaml:"pullPolicy,omitempty"`
		} `yaml:"image,omitempty"`
		PodSecurityPolicy struct {
			Create *bool `yaml:"create,omitempty"`
		} `yaml:"podSecurityPolicy,omitempty"`
		AdditionalPlugins []interface{} `yaml:"additionalPlugins,omitempty"`
		LogLevel          *string       `yaml:"logLevel,omitempty"`
		LogLevelFilter    *bool         `yaml:"logLevelFilter,omitempty"`
		VerifySsl         *bool         `yaml:"verifySsl,omitempty"`
		ProxyURI          *string       `yaml:"proxyUri,omitempty"`
		Compression       struct {
			Enabled  *bool   `yaml:"enabled,omitempty"`
			Encoding *string `yaml:"encoding,omitempty"`
		} `yaml:"compression,omitempty"`
		SecurityContext struct {
			FsGroup *int `yaml:"fsGroup,omitempty"`
		} `yaml:"securityContext,omitempty"`
		PodLabels      struct{} `yaml:"podLabels,omitempty"`
		PodAnnotations struct{} `yaml:"podAnnotations,omitempty"`
		ServiceLabels  struct{} `yaml:"serviceLabels,omitempty"`
		PvcLabels      struct{} `yaml:"pvcLabels,omitempty"`
		Persistence    struct {
			Enabled    *bool   `yaml:"enabled,omitempty"`
			AccessMode *string `yaml:"accessMode,omitempty"`
			Size       *string `yaml:"size,omitempty"`
		} `yaml:"persistence,omitempty"`
		Buffer struct {
			Type                *string `yaml:"type,omitempty"`
			FlushInterval       *string `yaml:"flushInterval,omitempty"`
			NumThreads          *int    `yaml:"numThreads,omitempty"`
			ChunkLimitSize      *string `yaml:"chunkLimitSize,omitempty"`
			QueueChunkLimitSize *int    `yaml:"queueChunkLimitSize,omitempty"`
			TotalLimitSize      *string `yaml:"totalLimitSize,omitempty"`
			RetryMaxInterval    *string `yaml:"retryMaxInterval,omitempty"`
			RetryForever        *bool   `yaml:"retryForever,omitempty"`
			Compress            *string `yaml:"compress,omitempty"`
			FilePaths           struct {
				Logs struct {
					Containers *string `yaml:"containers,omitempty"`
					Kubelet    *string `yaml:"kubelet,omitempty"`
					Systemd    *string `yaml:"systemd,omitempty"`
					Default    *string `yaml:"default,omitempty"`
				} `yaml:"logs,omitempty"`
				Metrics struct {
					Apiserver    *string `yaml:"apiserver,omitempty"`
					Kubelet      *string `yaml:"kubelet,omitempty"`
					Container    *string `yaml:"container,omitempty"`
					Controller   *string `yaml:"controller,omitempty"`
					Scheduler    *string `yaml:"scheduler,omitempty"`
					State        *string `yaml:"state,omitempty"`
					Node         *string `yaml:"node,omitempty"`
					ControlPlane *string `yaml:"control-plane,omitempty"`
					Default      *string `yaml:"default,omitempty"`
				} `yaml:"metrics,omitempty"`
				Events *string `yaml:"events,omitempty"`
				Traces *string `yaml:"traces,omitempty"`
			} `yaml:"filePaths,omitempty"`
			ExtraConf *string `yaml:"extraConf,omitempty"`
		} `yaml:"buffer,omitempty"`
		Monitoring struct {
			Input  *bool `yaml:"input,omitempty"`
			Output *bool `yaml:"output,omitempty"`
		} `yaml:"monitoring,omitempty"`
		Metadata struct {
			AnnotationMatch                   []*string `yaml:"annotation_match,omitempty"`
			CacheSize                         *string   `yaml:"cacheSize,omitempty"`
			CacheTTL                          *string   `yaml:"cacheTtl,omitempty"`
			CacheRefresh                      *string   `yaml:"cacheRefresh,omitempty"`
			CacheRefreshVariation             *string   `yaml:"cacheRefreshVariation,omitempty"`
			CacheRefreshApiserverRequestDelay *string   `yaml:"cacheRefreshApiserverRequestDelay,omitempty"`
			CacheRefreshExcludePodRegex       *string   `yaml:"cacheRefreshExcludePodRegex,omitempty"`
			PluginLogLevel                    *string   `yaml:"pluginLogLevel,omitempty"`
			CoreAPIVersions                   []*string `yaml:"coreApiVersions,omitempty"`
			APIGroups                         []*string `yaml:"apiGroups,omitempty"`
			AddOwners                         *bool     `yaml:"addOwners,omitempty"`
			AddService                        *bool     `yaml:"addService,omitempty"`
			APIServerURL                      *string   `yaml:"apiServerUrl,omitempty"`
		} `yaml:"metadata,omitempty"`
		Logs struct {
			Enabled     *bool `yaml:"enabled,omitempty"`
			Statefulset struct {
				NodeSelector    struct{}      `yaml:"nodeSelector,omitempty"`
				Tolerations     []interface{} `yaml:"tolerations,omitempty"`
				Affinity        struct{}      `yaml:"affinity,omitempty"`
				PodAntiAffinity *string       `yaml:"podAntiAffinity,omitempty"`
				ReplicaCount    *int          `yaml:"replicaCount,omitempty"`
				Resources       struct {
					Limits struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				PriorityClassName interface{} `yaml:"priorityClassName,omitempty"`
				PodLabels         struct{}    `yaml:"podLabels,omitempty"`
				PodAnnotations    struct{}    `yaml:"podAnnotations,omitempty"`
				Containers        struct {
					Fluentd struct {
						SecurityContext struct{} `yaml:"securityContext,omitempty"`
					} `yaml:"fluentd,omitempty"`
				} `yaml:"containers,omitempty"`
				InitContainers []interface{} `yaml:"initContainers,omitempty"`
			} `yaml:"statefulset,omitempty"`
			Autoscaling struct {
				Enabled                        *bool `yaml:"enabled,omitempty"`
				MinReplicas                    *int  `yaml:"minReplicas,omitempty"`
				MaxReplicas                    *int  `yaml:"maxReplicas,omitempty"`
				TargetCPUUtilizationPercentage *int  `yaml:"targetCPUUtilizationPercentage,omitempty"`
			} `yaml:"autoscaling,omitempty"`
			PodDisruptionBudget struct {
				MinAvailable *int `yaml:"minAvailable,omitempty"`
			} `yaml:"podDisruptionBudget,omitempty"`
			RawConfig *string `yaml:"rawConfig,omitempty"`
			Input     struct {
				ForwardExtraConf *string `yaml:"forwardExtraConf,omitempty"`
			} `yaml:"input,omitempty"`
			Output struct {
				LogFormat      *string `yaml:"logFormat,omitempty"`
				AddTimestamp   *bool   `yaml:"addTimestamp,omitempty"`
				TimestampKey   *string `yaml:"timestampKey,omitempty"`
				PluginLogLevel *string `yaml:"pluginLogLevel,omitempty"`
				ExtraConf      *string `yaml:"extraConf,omitempty"`
			} `yaml:"output,omitempty"`
			ExtraLogs  *string `yaml:"extraLogs,omitempty"`
			Containers struct {
				OverrideRawConfig              *string       `yaml:"overrideRawConfig,omitempty"`
				OutputConf                     *string       `yaml:"outputConf,omitempty"`
				OverrideOutputConf             *string       `yaml:"overrideOutputConf,omitempty"`
				SourceName                     *string       `yaml:"sourceName,omitempty"`
				SourceCategory                 *string       `yaml:"sourceCategory,omitempty"`
				SourceCategoryPrefix           *string       `yaml:"sourceCategoryPrefix,omitempty"`
				SourceCategoryReplaceDash      *string       `yaml:"sourceCategoryReplaceDash,omitempty"`
				ExcludeContainerRegex          *string       `yaml:"excludeContainerRegex,omitempty"`
				ExcludeHostRegex               *string       `yaml:"excludeHostRegex,omitempty"`
				ExcludeNamespaceRegex          *string       `yaml:"excludeNamespaceRegex,omitempty"`
				ExcludePodRegex                *string       `yaml:"excludePodRegex,omitempty"`
				PerContainerAnnotationsEnabled *bool         `yaml:"perContainerAnnotationsEnabled,omitempty"`
				PerContainerAnnotationPrefixes []interface{} `yaml:"perContainerAnnotationPrefixes,omitempty"`
				K8SMetadataFilter              struct {
					Watch               *string `yaml:"watch,omitempty"`
					CaFile              *string `yaml:"caFile,omitempty"`
					VerifySsl           *bool   `yaml:"verifySsl,omitempty"`
					ClientCert          *string `yaml:"clientCert,omitempty"`
					ClientKey           *string `yaml:"clientKey,omitempty"`
					BearerTokenFile     *string `yaml:"bearerTokenFile,omitempty"`
					TagToMetadataRegexp *string `yaml:"tagToMetadataRegexp,omitempty"`
				} `yaml:"k8sMetadataFilter,omitempty"`
				ExtraFilterPluginConf *string `yaml:"extraFilterPluginConf,omitempty"`
				ExtraOutputPluginConf *string `yaml:"extraOutputPluginConf,omitempty"`
				Multiline             struct {
					Enabled *bool `yaml:"enabled,omitempty"`
				} `yaml:"multiline,omitempty"`
			} `yaml:"containers,omitempty"`
			Kubelet struct {
				Enabled                   *bool   `yaml:"enabled,omitempty"`
				ExtraFilterPluginConf     *string `yaml:"extraFilterPluginConf,omitempty"`
				ExtraOutputPluginConf     *string `yaml:"extraOutputPluginConf,omitempty"`
				OutputConf                *string `yaml:"outputConf,omitempty"`
				OverrideOutputConf        *string `yaml:"overrideOutputConf,omitempty"`
				SourceName                *string `yaml:"sourceName,omitempty"`
				SourceCategory            *string `yaml:"sourceCategory,omitempty"`
				SourceCategoryPrefix      *string `yaml:"sourceCategoryPrefix,omitempty"`
				SourceCategoryReplaceDash *string `yaml:"sourceCategoryReplaceDash,omitempty"`
				ExcludeFacilityRegex      *string `yaml:"excludeFacilityRegex,omitempty"`
				ExcludeHostRegex          *string `yaml:"excludeHostRegex,omitempty"`
				ExcludePriorityRegex      *string `yaml:"excludePriorityRegex,omitempty"`
				ExcludeUnitRegex          *string `yaml:"excludeUnitRegex,omitempty"`
			} `yaml:"kubelet,omitempty"`
			Systemd struct {
				Enabled                   *bool   `yaml:"enabled,omitempty"`
				ExtraFilterPluginConf     *string `yaml:"extraFilterPluginConf,omitempty"`
				ExtraOutputPluginConf     *string `yaml:"extraOutputPluginConf,omitempty"`
				OutputConf                *string `yaml:"outputConf,omitempty"`
				OverrideOutputConf        *string `yaml:"overrideOutputConf,omitempty"`
				SourceName                *string `yaml:"sourceName,omitempty"`
				SourceCategory            *string `yaml:"sourceCategory,omitempty"`
				SourceCategoryPrefix      *string `yaml:"sourceCategoryPrefix,omitempty"`
				SourceCategoryReplaceDash *string `yaml:"sourceCategoryReplaceDash,omitempty"`
				ExcludeFacilityRegex      *string `yaml:"excludeFacilityRegex,omitempty"`
				ExcludeHostRegex          *string `yaml:"excludeHostRegex,omitempty"`
				ExcludePriorityRegex      *string `yaml:"excludePriorityRegex,omitempty"`
				ExcludeUnitRegex          *string `yaml:"excludeUnitRegex,omitempty"`
			} `yaml:"systemd,omitempty"`
			Default struct {
				ExtraFilterPluginConf     *string `yaml:"extraFilterPluginConf,omitempty"`
				ExtraOutputPluginConf     *string `yaml:"extraOutputPluginConf,omitempty"`
				OutputConf                *string `yaml:"outputConf,omitempty"`
				OverrideOutputConf        *string `yaml:"overrideOutputConf,omitempty"`
				SourceName                *string `yaml:"sourceName,omitempty"`
				SourceCategory            *string `yaml:"sourceCategory,omitempty"`
				SourceCategoryPrefix      *string `yaml:"sourceCategoryPrefix,omitempty"`
				SourceCategoryReplaceDash *string `yaml:"sourceCategoryReplaceDash,omitempty"`
				ExcludeFacilityRegex      *string `yaml:"excludeFacilityRegex,omitempty"`
				ExcludeHostRegex          *string `yaml:"excludeHostRegex,omitempty"`
				ExcludePriorityRegex      *string `yaml:"excludePriorityRegex,omitempty"`
				ExcludeUnitRegex          *string `yaml:"excludeUnitRegex,omitempty"`
			} `yaml:"default,omitempty"`
		} `yaml:"logs,omitempty"`
		Metrics struct {
			Enabled     *bool `yaml:"enabled,omitempty"`
			Statefulset struct {
				NodeSelector    struct{}      `yaml:"nodeSelector,omitempty"`
				Tolerations     []interface{} `yaml:"tolerations,omitempty"`
				Affinity        struct{}      `yaml:"affinity,omitempty"`
				PodAntiAffinity *string       `yaml:"podAntiAffinity,omitempty"`
				ReplicaCount    *int          `yaml:"replicaCount,omitempty"`
				Resources       struct {
					Limits struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				PriorityClassName interface{} `yaml:"priorityClassName,omitempty"`
				PodLabels         struct{}    `yaml:"podLabels,omitempty"`
				PodAnnotations    struct{}    `yaml:"podAnnotations,omitempty"`
				Containers        struct {
					Fluentd struct {
						SecurityContext struct{} `yaml:"securityContext,omitempty"`
					} `yaml:"fluentd,omitempty"`
				} `yaml:"containers,omitempty"`
				InitContainers []interface{} `yaml:"initContainers,omitempty"`
			} `yaml:"statefulset,omitempty"`
			Autoscaling struct {
				Enabled                        *bool `yaml:"enabled,omitempty"`
				MinReplicas                    *int  `yaml:"minReplicas,omitempty"`
				MaxReplicas                    *int  `yaml:"maxReplicas,omitempty"`
				TargetCPUUtilizationPercentage *int  `yaml:"targetCPUUtilizationPercentage,omitempty"`
			} `yaml:"autoscaling,omitempty"`
			PodDisruptionBudget struct {
				MinAvailable *int `yaml:"minAvailable,omitempty"`
			} `yaml:"podDisruptionBudget,omitempty"`
			RawConfig             *string `yaml:"rawConfig,omitempty"`
			OutputConf            *string `yaml:"outputConf,omitempty"`
			OverrideOutputConf    *string `yaml:"overrideOutputConf,omitempty"`
			ExtraFilterPluginConf *string `yaml:"extraFilterPluginConf,omitempty"`
			ExtraOutputPluginConf *string `yaml:"extraOutputPluginConf,omitempty"`
			Output                struct {
				Apiserver struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"apiserver,omitempty"`
				Kubelet struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"kubelet,omitempty"`
				Container struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Source *string `yaml:"source,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"container,omitempty"`
				Controller struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"controller,omitempty"`
				Scheduler struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"scheduler,omitempty"`
				State struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"state,omitempty"`
				Node struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"node,omitempty"`
				ControlPlane struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"control-plane,omitempty"`
				Default struct {
					Tag    *string `yaml:"tag,omitempty"`
					ID     *string `yaml:"id,omitempty"`
					Weight *int    `yaml:"weight,omitempty"`
				} `yaml:"default,omitempty"`
			} `yaml:"output,omitempty"`
		} `yaml:"metrics,omitempty"`
		Events struct {
			Enabled     *bool `yaml:"enabled,omitempty"`
			Statefulset struct {
				NodeSelector struct{}      `yaml:"nodeSelector,omitempty"`
				Tolerations  []interface{} `yaml:"tolerations,omitempty"`
				Affinity     struct{}      `yaml:"affinity,omitempty"`
				Resources    struct {
					Limits struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				PriorityClassName interface{} `yaml:"priorityClassName,omitempty"`
				PodLabels         struct{}    `yaml:"podLabels,omitempty"`
				PodAnnotations    struct{}    `yaml:"podAnnotations,omitempty"`
				Containers        struct {
					Fluentd struct {
						SecurityContext struct{} `yaml:"securityContext,omitempty"`
					} `yaml:"fluentd,omitempty"`
				} `yaml:"containers,omitempty"`
				InitContainers []interface{} `yaml:"initContainers,omitempty"`
			} `yaml:"statefulset,omitempty"`
			OverrideOutputConf *string `yaml:"overrideOutputConf,omitempty"`
			SourceCategory     *string `yaml:"sourceCategory,omitempty"`
		} `yaml:"events,omitempty"`
	} `yaml:"fluentd,omitempty"`
	MetricsServer struct {
		Enabled    *bool `yaml:"enabled,omitempty"`
		APIService struct {
			Create *bool `yaml:"create,omitempty"`
		} `yaml:"apiService,omitempty"`
		ExtraArgs struct {
			KubeletInsecureTLS           *bool   `yaml:"kubelet-insecure-tls,omitempty"`
			KubeletPreferredAddressTypes *string `yaml:"kubelet-preferred-address-types,omitempty"`
		} `yaml:"extraArgs,omitempty"`
	} `yaml:"metrics-server,omitempty"`
	FluentBit struct {
		Image struct {
			Repository *string `yaml:"repository,omitempty"`
			Tag        *string `yaml:"tag,omitempty"`
			PullPolicy *string `yaml:"pullPolicy,omitempty"`
		} `yaml:"image,omitempty"`
		Resources struct{} `yaml:"resources,omitempty"`
		Service   struct {
			Labels struct {
				SumologicComScrape *string `yaml:"sumologic.com/scrape,omitempty"`
			} `yaml:"labels,omitempty"`
		} `yaml:"service,omitempty"`
		PodLabels       struct{} `yaml:"podLabels,omitempty"`
		PodAnnotations  struct{} `yaml:"podAnnotations,omitempty"`
		SecurityContext struct {
			Capabilities struct {
				Drop []*string `yaml:"drop,omitempty"`
			} `yaml:"capabilities,omitempty"`
		} `yaml:"securityContext,omitempty"`
		Env []struct {
			Name      *string `yaml:"name,omitempty"`
			ValueFrom struct {
				ConfigMapKeyRef struct {
					Name *string `yaml:"name,omitempty"`
					Key  *string `yaml:"key,omitempty"`
				} `yaml:"configMapKeyRef,omitempty"`
				FieldRef struct {
					FieldPath *string `yaml:"fieldPath,omitempty"`
				} `yaml:"fieldRef,omitempty"`
			} `yaml:"valueFrom,omitempty"`
		} `yaml:"env,omitempty"`
		Tolerations []struct {
			Effect   *string `yaml:"effect,omitempty"`
			Operator *string `yaml:"operator,omitempty"`
		} `yaml:"tolerations,omitempty"`
		ExtraVolumeMounts []struct {
			MountPath *string `yaml:"mountPath,omitempty"`
			Name      *string `yaml:"name,omitempty"`
		} `yaml:"extraVolumeMounts,omitempty"`
		ExtraVolumes []struct {
			HostPath struct {
				Path *string `yaml:"path,omitempty"`
				Type *string `yaml:"type,omitempty"`
			} `yaml:"hostPath,omitempty"`
			Name *string `yaml:"name,omitempty"`
		} `yaml:"extraVolumes,omitempty"`
		Config struct {
			Service       *string `yaml:"service,omitempty"`
			Inputs        *string `yaml:"inputs,omitempty"`
			Outputs       *string `yaml:"outputs,omitempty"`
			CustomParsers *string `yaml:"customParsers,omitempty"`
		} `yaml:"config,omitempty"`
	} `yaml:"fluent-bit,omitempty"`
	KubePrometheusStack struct {
		CommonLabels                 struct{} `yaml:"commonLabels,omitempty"`
		AdditionalPrometheusRulesMap struct {
			Pre114NodeRules struct {
				Groups []struct {
					Name  *string `yaml:"name,omitempty"`
					Rules []struct {
						Expr   *string `yaml:"expr,omitempty"`
						Record *string `yaml:"record,omitempty"`
					} `yaml:"rules,omitempty"`
				} `yaml:"groups,omitempty"`
			} `yaml:"pre-1.14-node-rules,omitempty"`
		} `yaml:"additionalPrometheusRulesMap,omitempty"`
		KubeAPIServer struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"kubeApiServer,omitempty"`
		Kubelet struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				Probes            *bool       `yaml:"probes,omitempty"`
				Resource          *bool       `yaml:"resource,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
				CAdvisorMetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty,omitempty"`
					Replacement  *string   `yaml:"replacement,omitempty,omitempty"`
				} `yaml:"cAdvisorMetricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"kubelet,omitempty"`
		KubeControllerManager struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"kubeControllerManager,omitempty"`
		CoreDNS struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"coreDns,omitempty"`
		KubeEtcd struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"kubeEtcd,omitempty"`
		KubeScheduler struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"kubeScheduler,omitempty"`
		KubeStateMetrics struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"kubeStateMetrics,omitempty"`
		NodeExporter struct {
			ServiceMonitor struct {
				Interval          interface{} `yaml:"interval,omitempty"`
				MetricRelabelings []struct {
					Action       *string   `yaml:"action,omitempty"`
					Regex        *string   `yaml:"regex,omitempty"`
					SourceLabels []*string `yaml:"sourceLabels,omitempty"`
				} `yaml:"metricRelabelings,omitempty"`
			} `yaml:"serviceMonitor,omitempty"`
		} `yaml:"nodeExporter,omitempty"`
		Alertmanager struct {
			Enabled *bool `yaml:"enabled,omitempty"`
		} `yaml:"alertmanager,omitempty"`
		Grafana struct {
			Enabled                  *bool `yaml:"enabled,omitempty"`
			DefaultDashboardsEnabled *bool `yaml:"defaultDashboardsEnabled,omitempty"`
		} `yaml:"grafana,omitempty"`
		PrometheusOperator struct {
			PodLabels         struct{} `yaml:"podLabels,omitempty"`
			PodAnnotations    struct{} `yaml:"podAnnotations,omitempty"`
			Resources         struct{} `yaml:"resources,omitempty"`
			AdmissionWebhooks struct {
				Enabled *bool `yaml:"enabled,omitempty"`
			} `yaml:"admissionWebhooks,omitempty"`
			TLS struct {
				Enabled *bool `yaml:"enabled,omitempty"`
			} `yaml:"tls,omitempty"`
		} `yaml:"prometheusOperator,omitempty"`
		KubeStateMetricsSub struct {
			Image struct {
				Repository *string `yaml:"repository,omitempty"`
				Tag        *string `yaml:"tag,omitempty"`
			} `yaml:"image,omitempty"`
			CustomLabels   struct{} `yaml:"customLabels,omitempty"`
			PodAnnotations struct{} `yaml:"podAnnotations,omitempty"`
			Resources      struct{} `yaml:"resources,omitempty"`
		} `yaml:"kube-state-metrics,omitempty"`
		PrometheusNodeExporter struct {
			Image struct {
				Tag *string `yaml:"tag,omitempty"`
			} `yaml:"image,omitempty"`
			PodLabels      struct{} `yaml:"podLabels,omitempty"`
			PodAnnotations struct{} `yaml:"podAnnotations,omitempty"`
			Resources      struct{} `yaml:"resources,omitempty"`
		} `yaml:"prometheus-node-exporter,omitempty"`
		Prometheus struct {
			AdditionalServiceMonitors []struct {
				Name             *string `yaml:"name,omitempty"`
				AdditionalLabels struct {
					SumologicComApp *string `yaml:"sumologic.com/app,omitempty"`
				} `yaml:"additionalLabels,omitempty"`
				Endpoints []struct {
					Port *string `yaml:"port,omitempty"`
				} `yaml:"endpoints,omitempty"`
				NamespaceSelector struct {
					MatchNames []*string `yaml:"matchNames,omitempty"`
				} `yaml:"namespaceSelector,omitempty"`
				Selector struct {
					MatchLabels struct {
						SumologicComApp    *string `yaml:"sumologic.com/app,omitempty"`
						SumologicComScrape *string `yaml:"sumologic.com/scrape,omitempty"`
					} `yaml:"matchLabels,omitempty"`
				} `yaml:"selector,omitempty"`
			} `yaml:"additionalServiceMonitors,omitempty"`
			PrometheusSpec struct {
				ScrapeInterval *string `yaml:"scrapeInterval,omitempty"`
				Retention      *string `yaml:"retention,omitempty"`
				PodMetadata    struct {
					Labels      struct{} `yaml:"labels,omitempty"`
					Annotations struct{} `yaml:"annotations,omitempty"`
				} `yaml:"podMetadata,omitempty"`
				Resources struct {
					Limits struct {
						CPU    *string `yaml:"cpu,omitempty"`
						Memory *string `yaml:"memory,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						CPU    *string `yaml:"cpu,omitempty"`
						Memory *string `yaml:"memory,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				Thanos struct {
					BaseImage *string `yaml:"baseImage,omitempty"`
					Version   *string `yaml:"version,omitempty"`
					Resources struct {
						Limits struct {
							CPU    *string `yaml:"cpu,omitempty"`
							Memory *string `yaml:"memory,omitempty"`
						} `yaml:"limits,omitempty"`
						Requests struct {
							CPU    *string `yaml:"cpu,omitempty"`
							Memory *string `yaml:"memory,omitempty"`
						} `yaml:"requests,omitempty"`
					} `yaml:"resources,omitempty"`
				} `yaml:"thanos,omitempty"`
				Containers []struct {
					Name *string `yaml:"name,omitempty"`
					Env  []struct {
						Name      *string `yaml:"name,omitempty"`
						ValueFrom struct {
							ConfigMapKeyRef struct {
								Name *string `yaml:"name,omitempty"`
								Key  *string `yaml:"key,omitempty"`
							} `yaml:"configMapKeyRef,omitempty"`
						} `yaml:"valueFrom,omitempty"`
					} `yaml:"env,omitempty"`
				} `yaml:"containers,omitempty"`
				WalCompression          *bool `yaml:"walCompression,omitempty"`
				AdditionalScrapeConfigs []struct {
					JobName             *string `yaml:"job_name,omitempty"`
					KubernetesSdConfigs []struct {
						Role *string `yaml:"role,omitempty"`
					} `yaml:"kubernetes_sd_configs,omitempty"`
					RelabelConfigs []struct {
						SourceLabels []*string `yaml:"source_labels,omitempty,omitempty"`
						Action       *string   `yaml:"action,omitempty"`
						Regex        *string   `yaml:"regex,omitempty,omitempty"`
						TargetLabel  *string   `yaml:"target_label,omitempty,omitempty"`
						Replacement  *string   `yaml:"replacement,omitempty,omitempty"`
						Separator    *string   `yaml:"separator,omitempty,omitempty"`
					} `yaml:"relabel_configs,omitempty"`
				} `yaml:"additionalScrapeConfigs,omitempty"`
				RemoteWrite []struct {
					URL                 *string `yaml:"url,omitempty"`
					RemoteTimeout       *string `yaml:"remoteTimeout,omitempty"`
					WriteRelabelConfigs []struct {
						Action       *string   `yaml:"action,omitempty"`
						Regex        *string   `yaml:"regex,omitempty"`
						SourceLabels []*string `yaml:"sourceLabels,omitempty,omitempty"`
						Replacement  *string   `yaml:"replacement,omitempty,omitempty"`
					} `yaml:"writeRelabelConfigs,omitempty"`
				} `yaml:"remoteWrite,omitempty"`
			} `yaml:"prometheusSpec,omitempty"`
		} `yaml:"prometheus,omitempty"`
	} `yaml:"kube-prometheus-stack,omitempty"`
	Otelagent struct {
		Enabled   *bool `yaml:"enabled,omitempty"`
		Daemonset struct {
			NodeSelector struct{}      `yaml:"nodeSelector,omitempty"`
			Tolerations  []interface{} `yaml:"tolerations,omitempty"`
			Resources    struct {
				Limits struct {
					Memory *string `yaml:"memory,omitempty"`
					CPU    *string `yaml:"cpu,omitempty"`
				} `yaml:"limits,omitempty"`
				Requests struct {
					Memory *string `yaml:"memory,omitempty"`
					CPU    *string `yaml:"cpu,omitempty"`
				} `yaml:"requests,omitempty"`
			} `yaml:"resources,omitempty"`
			PodLabels      struct{} `yaml:"podLabels,omitempty"`
			PodAnnotations struct{} `yaml:"podAnnotations,omitempty"`
			Image          struct {
				Repository *string `yaml:"repository,omitempty"`
				Tag        *string `yaml:"tag,omitempty"`
				PullPolicy *string `yaml:"pullPolicy,omitempty"`
			} `yaml:"image,omitempty"`
		} `yaml:"daemonset,omitempty"`
		Config struct {
			Receivers struct {
				Jaeger struct {
					Protocols struct {
						ThriftCompact struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"thrift_compact,omitempty"`
						ThriftBinary struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"thrift_binary,omitempty"`
						Grpc struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"grpc,omitempty"`
						ThriftHTTP struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"thrift_http,omitempty"`
					} `yaml:"protocols,omitempty"`
				} `yaml:"jaeger,omitempty"`
				Opencensus struct {
					Endpoint *string `yaml:"endpoint,omitempty"`
				} `yaml:"opencensus,omitempty"`
				Otlp struct {
					Protocols struct {
						Grpc struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"grpc,omitempty"`
						HTTP struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"http,omitempty"`
					} `yaml:"protocols,omitempty"`
				} `yaml:"otlp,omitempty"`
				Zipkin struct {
					Endpoint *string `yaml:"endpoint,omitempty"`
				} `yaml:"zipkin,omitempty"`
			} `yaml:"receivers,omitempty"`
			Processors struct {
				K8STagger struct {
					Passthrough *bool `yaml:"passthrough,omitempty"`
				} `yaml:"k8s_tagger,omitempty"`
				MemoryLimiter struct {
					CheckInterval *string `yaml:"check_interval,omitempty"`
					LimitMib      *int    `yaml:"limit_mib,omitempty"`
				} `yaml:"memory_limiter,omitempty"`
				Batch struct {
					SendBatchSize *int    `yaml:"send_batch_size,omitempty"`
					Timeout       *string `yaml:"timeout,omitempty"`
				} `yaml:"batch,omitempty"`
			} `yaml:"processors,omitempty"`
			Extensions struct {
				HealthCheck   struct{} `yaml:"health_check,omitempty"`
				MemoryBallast struct {
					SizeMib *int `yaml:"size_mib,omitempty"`
				} `yaml:"memory_ballast,omitempty"`
			} `yaml:"extensions,omitempty"`
			Exporters struct {
				Otlp struct {
					Endpoint *string `yaml:"endpoint,omitempty"`
					TLS      struct {
						Insecure *bool `yaml:"insecure,omitempty"`
					} `yaml:"tls,omitempty"`
				} `yaml:"otlp,omitempty"`
			} `yaml:"exporters,omitempty"`
			Service struct {
				Extensions []*string `yaml:"extensions,omitempty"`
				Pipelines  struct {
					Traces struct {
						Receivers  []*string `yaml:"receivers,omitempty"`
						Processors []*string `yaml:"processors,omitempty"`
						Exporters  []*string `yaml:"exporters,omitempty"`
					} `yaml:"traces,omitempty"`
					Metrics struct {
						Receivers  []*string `yaml:"receivers,omitempty"`
						Processors []*string `yaml:"processors,omitempty"`
						Exporters  []*string `yaml:"exporters,omitempty"`
					} `yaml:"metrics,omitempty"`
				} `yaml:"pipelines,omitempty"`
			} `yaml:"service,omitempty"`
		} `yaml:"config,omitempty"`
	} `yaml:"otelagent,omitempty"`
	Otelcol struct {
		Deployment struct {
			NodeSelector struct{}      `yaml:"nodeSelector,omitempty"`
			Tolerations  []interface{} `yaml:"tolerations,omitempty"`
			Replicas     *int          `yaml:"replicas,omitempty"`
			Resources    struct {
				Limits struct {
					Memory *string `yaml:"memory,omitempty"`
					CPU    *string `yaml:"cpu,omitempty"`
				} `yaml:"limits,omitempty"`
				Requests struct {
					Memory *string `yaml:"memory,omitempty"`
					CPU    *string `yaml:"cpu,omitempty"`
				} `yaml:"requests,omitempty"`
			} `yaml:"resources,omitempty"`
			PriorityClassName interface{} `yaml:"priorityClassName,omitempty"`
			PodLabels         struct{}    `yaml:"podLabels,omitempty"`
			PodAnnotations    struct{}    `yaml:"podAnnotations,omitempty"`
			Image             struct {
				Repository *string `yaml:"repository,omitempty"`
				Tag        *string `yaml:"tag,omitempty"`
				PullPolicy *string `yaml:"pullPolicy,omitempty"`
			} `yaml:"image,omitempty"`
		} `yaml:"deployment,omitempty"`
		LogLevelFilter *bool `yaml:"logLevelFilter,omitempty"`
		Metrics        struct {
			Enabled *bool `yaml:"enabled,omitempty"`
		} `yaml:"metrics,omitempty"`
		Config struct {
			Receivers struct {
				Jaeger struct {
					Protocols struct {
						ThriftCompact struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"thrift_compact,omitempty"`
						ThriftBinary struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"thrift_binary,omitempty"`
						Grpc struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"grpc,omitempty"`
						ThriftHTTP struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"thrift_http,omitempty"`
					} `yaml:"protocols,omitempty"`
				} `yaml:"jaeger,omitempty"`
				Opencensus struct {
					Endpoint *string `yaml:"endpoint,omitempty"`
				} `yaml:"opencensus,omitempty"`
				Otlp struct {
					Protocols struct {
						Grpc struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"grpc,omitempty"`
						HTTP struct {
							Endpoint *string `yaml:"endpoint,omitempty"`
						} `yaml:"http,omitempty"`
					} `yaml:"protocols,omitempty"`
				} `yaml:"otlp,omitempty"`
				Zipkin struct {
					Endpoint *string `yaml:"endpoint,omitempty"`
				} `yaml:"zipkin,omitempty"`
			} `yaml:"receivers,omitempty"`
			Processors struct {
				K8STagger struct {
					Passthrough        *bool `yaml:"passthrough,omitempty"`
					OwnerLookupEnabled *bool `yaml:"owner_lookup_enabled,omitempty"`
					Extract            struct {
						Metadata    []*string `yaml:"metadata,omitempty"`
						Annotations []struct {
							TagName *string `yaml:"tag_name,omitempty"`
							Key     *string `yaml:"key,omitempty"`
						} `yaml:"annotations,omitempty"`
						NamespaceLabels []struct {
							TagName *string `yaml:"tag_name,omitempty"`
							Key     *string `yaml:"key,omitempty"`
						} `yaml:"namespace_labels,omitempty"`
						Labels []struct {
							TagName *string `yaml:"tag_name,omitempty"`
							Key     *string `yaml:"key,omitempty"`
						} `yaml:"labels,omitempty"`
					} `yaml:"extract,omitempty"`
				} `yaml:"k8s_tagger,omitempty"`
				Source struct {
					Collector                 *string `yaml:"collector,omitempty"`
					SourceName                *string `yaml:"source_name,omitempty"`
					SourceCategory            *string `yaml:"source_category,omitempty"`
					SourceCategoryPrefix      *string `yaml:"source_category_prefix,omitempty"`
					SourceCategoryReplaceDash *string `yaml:"source_category_replace_dash,omitempty"`
					ExcludeNamespaceRegex     *string `yaml:"exclude_namespace_regex,omitempty"`
					ExcludePodRegex           *string `yaml:"exclude_pod_regex,omitempty"`
					ExcludeContainerRegex     *string `yaml:"exclude_container_regex,omitempty"`
					ExcludeHostRegex          *string `yaml:"exclude_host_regex,omitempty"`
					AnnotationPrefix          *string `yaml:"annotation_prefix,omitempty"`
					PodTemplateHashKey        *string `yaml:"pod_template_hash_key,omitempty"`
					PodNameKey                *string `yaml:"pod_name_key,omitempty"`
					NamespaceKey              *string `yaml:"namespace_key,omitempty"`
					PodKey                    *string `yaml:"pod_key,omitempty"`
					ContainerKey              *string `yaml:"container_key,omitempty"`
					SourceHostKey             *string `yaml:"source_host_key,omitempty"`
				} `yaml:"source,omitempty"`
				Resource struct {
					Attributes []struct {
						Key    *string `yaml:"key,omitempty"`
						Value  *string `yaml:"value,omitempty"`
						Action *string `yaml:"action,omitempty"`
					} `yaml:"attributes,omitempty"`
				} `yaml:"resource,omitempty"`
				MemoryLimiter struct {
					CheckInterval *string `yaml:"check_interval,omitempty"`
					LimitMib      *int    `yaml:"limit_mib,omitempty"`
				} `yaml:"memory_limiter,omitempty"`
				CascadingFilter struct {
					NumTraces *int `yaml:"num_traces,omitempty"`
				} `yaml:"cascading_filter,omitempty"`
				Batch struct {
					SendBatchSize    *int    `yaml:"send_batch_size,omitempty"`
					SendBatchMaxSize *int    `yaml:"send_batch_max_size,omitempty"`
					Timeout          *string `yaml:"timeout,omitempty"`
				} `yaml:"batch,omitempty"`
				Resourcedetection struct {
					Detectors []*string `yaml:"detectors,omitempty"`
					Timeout   *string   `yaml:"timeout,omitempty"`
					Override  *bool     `yaml:"override,omitempty"`
				} `yaml:"resourcedetection,omitempty"`
			} `yaml:"processors,omitempty"`
			Extensions struct {
				HealthCheck   struct{} `yaml:"health_check,omitempty"`
				MemoryBallast struct {
					SizeMib *int `yaml:"size_mib,omitempty"`
				} `yaml:"memory_ballast,omitempty"`
			} `yaml:"extensions,omitempty"`
			Exporters struct {
				Zipkin struct {
					Endpoint *string `yaml:"endpoint,omitempty"`
				} `yaml:"zipkin,omitempty"`
				Otlphttp struct {
					TracesEndpoint *string `yaml:"traces_endpoint,omitempty"`
					Compression    *string `yaml:"compression,omitempty"`
				} `yaml:"otlphttp,omitempty"`
				Sumologic struct {
					Endpoint           *string   `yaml:"endpoint,omitempty"`
					CompressEncoding   *string   `yaml:"compress_encoding,omitempty"`
					MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
					LogFormat          *string   `yaml:"log_format,omitempty"`
					MetricFormat       *string   `yaml:"metric_format,omitempty"`
					SourceCategory     *string   `yaml:"source_category,omitempty"`
					SourceName         *string   `yaml:"source_name,omitempty"`
					SourceHost         *string   `yaml:"source_host,omitempty"`
					MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					Timeout            *string   `yaml:"timeout,omitempty"`
					RetryOnFailure     struct {
						Enabled         *bool   `yaml:"enabled,omitempty"`
						InitialInterval *string `yaml:"initial_interval,omitempty"`
						MaxInterval     *string `yaml:"max_interval,omitempty"`
						MaxElapsedTime  *string `yaml:"max_elapsed_time,omitempty"`
					} `yaml:"retry_on_failure,omitempty"`
					SendingQueue struct {
						Enabled      *bool `yaml:"enabled,omitempty"`
						NumConsumers *int  `yaml:"num_consumers,omitempty"`
						QueueSize    *int  `yaml:"queue_size,omitempty"`
					} `yaml:"sending_queue,omitempty"`
				} `yaml:"sumologic,omitempty"`
			} `yaml:"exporters,omitempty"`
			Service struct {
				Extensions []*string `yaml:"extensions,omitempty"`
				Pipelines  struct {
					Traces struct {
						Receivers  []*string `yaml:"receivers,omitempty"`
						Processors []*string `yaml:"processors,omitempty"`
						Exporters  []*string `yaml:"exporters,omitempty"`
					} `yaml:"traces,omitempty"`
					Metrics struct {
						Receivers  []*string `yaml:"receivers,omitempty"`
						Processors []*string `yaml:"processors,omitempty"`
						Exporters  []*string `yaml:"exporters,omitempty"`
					} `yaml:"metrics,omitempty"`
				} `yaml:"pipelines,omitempty"`
			} `yaml:"service,omitempty"`
		} `yaml:"config,omitempty"`
	} `yaml:"otelcol,omitempty"`
	Metadata struct {
		Image struct {
			Repository *string `yaml:"repository,omitempty"`
			Tag        *string `yaml:"tag,omitempty"`
			PullPolicy *string `yaml:"pullPolicy,omitempty"`
		} `yaml:"image,omitempty"`
		SecurityContext struct {
			FsGroup *int `yaml:"fsGroup,omitempty"`
		} `yaml:"securityContext,omitempty"`
		PodLabels      struct{} `yaml:"podLabels,omitempty"`
		PodAnnotations struct{} `yaml:"podAnnotations,omitempty"`
		ServiceLabels  struct{} `yaml:"serviceLabels,omitempty"`
		Persistence    struct {
			Enabled    *bool    `yaml:"enabled,omitempty"`
			AccessMode *string  `yaml:"accessMode,omitempty"`
			Size       *string  `yaml:"size,omitempty"`
			PvcLabels  struct{} `yaml:"pvcLabels,omitempty"`
		} `yaml:"persistence,omitempty"`
		Metrics struct {
			Enabled  *bool   `yaml:"enabled,omitempty"`
			LogLevel *string `yaml:"logLevel,omitempty"`
			Config   struct {
				Receivers struct {
					Telegraf struct {
						AgentConfig *string `yaml:"agent_config,omitempty"`
					} `yaml:"telegraf,omitempty"`
				} `yaml:"receivers,omitempty"`
				Extensions struct {
					HealthCheck struct{} `yaml:"health_check,omitempty"`
					FileStorage struct {
						Directory *string `yaml:"directory,omitempty"`
						Timeout   *string `yaml:"timeout,omitempty"`
					} `yaml:"file_storage,omitempty"`
					Pprof struct{} `yaml:"pprof,omitempty"`
				} `yaml:"extensions,omitempty"`
				Exporters struct {
					SumologicDefault struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/default,omitempty"`
					SumologicApiserver struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/apiserver,omitempty"`
					SumologicControlPlane struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/control_plane,omitempty"`
					SumologicController struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/controller,omitempty"`
					SumologicKubelet struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/kubelet,omitempty"`
					SumologicNode struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/node,omitempty"`
					SumologicScheduler struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/scheduler,omitempty"`
					SumologicState struct {
						MetricFormat *string `yaml:"metric_format,omitempty"`
						Endpoint     *string `yaml:"endpoint,omitempty"`
						SendingQueue struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
							NumConsumers             *int    `yaml:"num_consumers,omitempty"`
							QueueSize                *int    `yaml:"queue_size,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MaxRequestBodySize *int      `yaml:"max_request_body_size,omitempty"`
						Timeout            *string   `yaml:"timeout,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/state,omitempty"`
				} `yaml:"exporters,omitempty"`
				Processors struct {
					Metricstransform struct {
						Transforms struct {
							Include   *string `yaml:"include,omitempty"`
							MatchType *string `yaml:"match_type,omitempty"`
							Action    *string `yaml:"action,omitempty"`
							NewName   *string `yaml:"new_name,omitempty"`
						} `yaml:"transforms,omitempty"`
					} `yaml:"metricstransform,omitempty"`
					ResourceDeleteSourceMetadata struct {
						Attributes []struct {
							Key    *string `yaml:"key,omitempty"`
							Action *string `yaml:"action,omitempty"`
						} `yaml:"attributes,omitempty"`
					} `yaml:"resource/delete_source_metadata,omitempty"`
					Resource struct {
						Attributes []struct {
							Key           *string `yaml:"key,omitempty"`
							FromAttribute *string `yaml:"from_attribute,omitempty,omitempty"`
							Action        *string `yaml:"action,omitempty"`
							Value         *string `yaml:"value,omitempty,omitempty"`
						} `yaml:"attributes,omitempty"`
					} `yaml:"resource,omitempty"`
					ResourceRemoveK8SPodPodName struct {
						Attributes []struct {
							Action *string `yaml:"action,omitempty"`
							Key    *string `yaml:"key,omitempty"`
						} `yaml:"attributes,omitempty"`
					} `yaml:"resource/remove_k8s_pod_pod_name,omitempty"`
					Routing struct {
						FromAttribute    *string   `yaml:"from_attribute,omitempty"`
						AttributeSource  *string   `yaml:"attribute_source,omitempty"`
						DefaultExporters []*string `yaml:"default_exporters,omitempty"`
						Table            []struct {
							Value     *string   `yaml:"value,omitempty"`
							Exporters []*string `yaml:"exporters,omitempty"`
						} `yaml:"table,omitempty"`
					} `yaml:"routing,omitempty"`
					MemoryLimiter struct {
						CheckInterval *string `yaml:"check_interval,omitempty"`
						LimitMib      *int    `yaml:"limit_mib,omitempty"`
					} `yaml:"memory_limiter,omitempty"`
					Batch struct {
						SendBatchSize *int    `yaml:"send_batch_size,omitempty"`
						Timeout       *string `yaml:"timeout,omitempty"`
					} `yaml:"batch,omitempty"`
					K8STagger struct {
						Passthrough        *bool `yaml:"passthrough,omitempty"`
						OwnerLookupEnabled *bool `yaml:"owner_lookup_enabled,omitempty"`
						Extract            struct {
							Metadata []*string `yaml:"metadata,omitempty"`
							Labels   []struct {
								TagName *string `yaml:"tag_name,omitempty"`
								Key     *string `yaml:"key,omitempty"`
							} `yaml:"labels,omitempty"`
							Delimiter *string `yaml:"delimiter,omitempty"`
						} `yaml:"extract,omitempty"`
						PodAssociation []struct {
							From *string `yaml:"from,omitempty"`
						} `yaml:"pod_association,omitempty"`
					} `yaml:"k8s_tagger,omitempty"`
					Source struct {
						Collector *string `yaml:"collector,omitempty"`
					} `yaml:"source,omitempty"`
				} `yaml:"processors,omitempty"`
				Service struct {
					Telemetry struct {
						Logs struct {
							Level *string `yaml:"level,omitempty"`
						} `yaml:"logs,omitempty"`
					} `yaml:"telemetry,omitempty"`
					Extensions []*string `yaml:"extensions,omitempty"`
					Pipelines  struct {
						Metrics struct {
							Receivers  []*string `yaml:"receivers,omitempty"`
							Processors []*string `yaml:"processors,omitempty"`
							Exporters  []*string `yaml:"exporters,omitempty"`
						} `yaml:"metrics,omitempty"`
					} `yaml:"pipelines,omitempty"`
				} `yaml:"service,omitempty"`
			} `yaml:"config,omitempty"`
			Statefulset struct {
				NodeSelector    struct{}      `yaml:"nodeSelector,omitempty"`
				Tolerations     []interface{} `yaml:"tolerations,omitempty"`
				Affinity        struct{}      `yaml:"affinity,omitempty"`
				PodAntiAffinity *string       `yaml:"podAntiAffinity,omitempty"`
				ReplicaCount    *int          `yaml:"replicaCount,omitempty"`
				Resources       struct {
					Limits struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				PriorityClassName interface{} `yaml:"priorityClassName,omitempty"`
				PodLabels         struct{}    `yaml:"podLabels,omitempty"`
				PodAnnotations    struct{}    `yaml:"podAnnotations,omitempty"`
				Containers        struct {
					Otelcol struct {
						SecurityContext struct{} `yaml:"securityContext,omitempty"`
						LivenessProbe   struct {
							InitialDelaySeconds *int `yaml:"initialDelaySeconds,omitempty"`
							PeriodSeconds       *int `yaml:"periodSeconds,omitempty"`
							TimeoutSeconds      *int `yaml:"timeoutSeconds,omitempty"`
							FailureThreshold    *int `yaml:"failureThreshold,omitempty"`
						} `yaml:"livenessProbe,omitempty"`
						ReadinessProbe struct {
							InitialDelaySeconds *int `yaml:"initialDelaySeconds,omitempty"`
							PeriodSeconds       *int `yaml:"periodSeconds,omitempty"`
							TimeoutSeconds      *int `yaml:"timeoutSeconds,omitempty"`
							FailureThreshold    *int `yaml:"failureThreshold,omitempty"`
						} `yaml:"readinessProbe,omitempty"`
						StartupProbe struct {
							PeriodSeconds    *int `yaml:"periodSeconds,omitempty"`
							FailureThreshold *int `yaml:"failureThreshold,omitempty"`
						} `yaml:"startupProbe,omitempty"`
					} `yaml:"otelcol,omitempty"`
				} `yaml:"containers,omitempty"`
			} `yaml:"statefulset,omitempty"`
			Autoscaling struct {
				Enabled                        *bool `yaml:"enabled,omitempty"`
				MinReplicas                    *int  `yaml:"minReplicas,omitempty"`
				MaxReplicas                    *int  `yaml:"maxReplicas,omitempty"`
				TargetCPUUtilizationPercentage *int  `yaml:"targetCPUUtilizationPercentage,omitempty"`
			} `yaml:"autoscaling,omitempty"`
			PodDisruptionBudget struct {
				MinAvailable *int `yaml:"minAvailable,omitempty"`
			} `yaml:"podDisruptionBudget,omitempty"`
		} `yaml:"metrics,omitempty"`
		Logs struct {
			Enabled  *bool   `yaml:"enabled,omitempty"`
			LogLevel *string `yaml:"logLevel,omitempty"`
			Config   struct {
				Receivers struct {
					Fluentforward struct {
						Endpoint *string `yaml:"endpoint,omitempty"`
					} `yaml:"fluentforward,omitempty"`
					Otlp struct {
						Protocols struct {
							HTTP struct {
								Endpoint *string `yaml:"endpoint,omitempty"`
							} `yaml:"http,omitempty"`
						} `yaml:"protocols,omitempty"`
					} `yaml:"otlp,omitempty"`
				} `yaml:"receivers,omitempty"`
				Extensions struct {
					HealthCheck struct{} `yaml:"health_check,omitempty"`
					FileStorage struct {
						Directory *string `yaml:"directory,omitempty"`
						Timeout   *string `yaml:"timeout,omitempty"`
					} `yaml:"file_storage,omitempty"`
					Pprof struct{} `yaml:"pprof,omitempty"`
				} `yaml:"extensions,omitempty"`
				Exporters struct {
					SumologicContainers struct {
						LogFormat *string `yaml:"log_format,omitempty"`
						JSONLogs  struct {
							AddTimestamp *bool   `yaml:"add_timestamp,omitempty"`
							TimestampKey *string `yaml:"timestamp_key,omitempty"`
						} `yaml:"json_logs,omitempty"`
						Endpoint       *string `yaml:"endpoint,omitempty"`
						SourceName     *string `yaml:"source_name,omitempty"`
						SourceCategory *string `yaml:"source_category,omitempty"`
						SourceHost     *string `yaml:"source_host,omitempty"`
						SendingQueue   struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/containers,omitempty"`
					SumologicSystemd struct {
						LogFormat *string `yaml:"log_format,omitempty"`
						JSONLogs  struct {
							AddTimestamp *bool   `yaml:"add_timestamp,omitempty"`
							TimestampKey *string `yaml:"timestamp_key,omitempty"`
						} `yaml:"json_logs,omitempty"`
						Endpoint       *string `yaml:"endpoint,omitempty"`
						SourceName     *string `yaml:"source_name,omitempty"`
						SourceCategory *string `yaml:"source_category,omitempty"`
						SourceHost     *string `yaml:"source_host,omitempty"`
						SendingQueue   struct {
							Enabled                  *bool   `yaml:"enabled,omitempty"`
							PersistentStorageEnabled *string `yaml:"persistent_storage_enabled,omitempty"`
						} `yaml:"sending_queue,omitempty"`
						MetadataAttributes []*string `yaml:"metadata_attributes,omitempty"`
					} `yaml:"sumologic/systemd,omitempty"`
				} `yaml:"exporters,omitempty"`
				Processors struct {
					AttributesRemoveFluentTag struct {
						Actions []struct {
							Action *string `yaml:"action,omitempty"`
							Key    *string `yaml:"key,omitempty"`
						} `yaml:"actions,omitempty"`
					} `yaml:"attributes/remove_fluent_tag,omitempty"`
					MemoryLimiter struct {
						CheckInterval *string `yaml:"check_interval,omitempty"`
						LimitMib      *int    `yaml:"limit_mib,omitempty"`
					} `yaml:"memory_limiter,omitempty"`
					Batch struct {
						SendBatchSize *int    `yaml:"send_batch_size,omitempty"`
						Timeout       *string `yaml:"timeout,omitempty"`
					} `yaml:"batch,omitempty"`
					FilterIncludeFluentTagContainers struct {
						Logs struct {
							Include struct {
								MatchType        *string `yaml:"match_type,omitempty"`
								RecordAttributes []struct {
									Key   *string `yaml:"key,omitempty"`
									Value *string `yaml:"value,omitempty"`
								} `yaml:"record_attributes,omitempty"`
							} `yaml:"include,omitempty"`
						} `yaml:"logs,omitempty"`
					} `yaml:"filter/include_fluent_tag_containers,omitempty"`
					AttributesContainers struct {
						Actions []struct {
							Action        *string `yaml:"action,omitempty"`
							Key           *string `yaml:"key,omitempty"`
							Pattern       *string `yaml:"pattern,omitempty,omitempty"`
							FromAttribute *string `yaml:"from_attribute,omitempty,omitempty"`
						} `yaml:"actions,omitempty"`
					} `yaml:"attributes/containers,omitempty"`
					ResourceContainersCopyNodeToHost struct {
						Attributes []struct {
							Action        *string `yaml:"action,omitempty"`
							Key           *string `yaml:"key,omitempty"`
							FromAttribute *string `yaml:"from_attribute,omitempty"`
						} `yaml:"attributes,omitempty"`
					} `yaml:"resource/containers_copy_node_to_host,omitempty"`
					GroupbyattrsContainers struct {
						Keys []*string `yaml:"keys,omitempty"`
					} `yaml:"groupbyattrs/containers,omitempty"`
					K8STagger struct {
						Passthrough        *bool `yaml:"passthrough,omitempty"`
						OwnerLookupEnabled *bool `yaml:"owner_lookup_enabled,omitempty"`
						Extract            struct {
							Metadata    []*string `yaml:"metadata,omitempty"`
							Annotations []struct {
								TagName *string `yaml:"tag_name,omitempty"`
								Key     *string `yaml:"key,omitempty"`
							} `yaml:"annotations,omitempty"`
							NamespaceLabels []struct {
								TagName *string `yaml:"tag_name,omitempty"`
								Key     *string `yaml:"key,omitempty"`
							} `yaml:"namespace_labels,omitempty"`
							Labels []struct {
								TagName *string `yaml:"tag_name,omitempty"`
								Key     *string `yaml:"key,omitempty"`
							} `yaml:"labels,omitempty"`
							Delimiter *string `yaml:"delimiter,omitempty"`
						} `yaml:"extract,omitempty"`
						PodAssociation []struct {
							From *string `yaml:"from,omitempty"`
						} `yaml:"pod_association,omitempty"`
					} `yaml:"k8s_tagger,omitempty"`
					SourceContainers struct {
						Collector                 *string `yaml:"collector,omitempty"`
						SourceHost                *string `yaml:"source_host,omitempty"`
						SourceName                *string `yaml:"source_name,omitempty"`
						SourceCategory            *string `yaml:"source_category,omitempty"`
						SourceCategoryPrefix      *string `yaml:"source_category_prefix,omitempty"`
						SourceCategoryReplaceDash *string `yaml:"source_category_replace_dash,omitempty"`
						Exclude                   struct {
							K8SNamespaceName *string `yaml:"k8s.namespace.name,omitempty"`
							K8SPodName       *string `yaml:"k8s.pod.name,omitempty"`
							K8SContainerName *string `yaml:"k8s.container.name,omitempty"`
							K8SPodHostname   *string `yaml:"k8s.pod.hostname,omitempty"`
						} `yaml:"exclude,omitempty"`
						AnnotationPrefix     *string `yaml:"annotation_prefix,omitempty"`
						PodTemplateHashKey   *string `yaml:"pod_template_hash_key,omitempty"`
						PodNameKey           *string `yaml:"pod_name_key,omitempty"`
						PodKey               *string `yaml:"pod_key,omitempty"`
						ContainerAnnotations struct {
							Enabled  *string `yaml:"enabled,omitempty"`
							Prefixes *string `yaml:"prefixes,omitempty"`
						} `yaml:"container_annotations,omitempty"`
					} `yaml:"source/containers,omitempty"`
					FilterIncludeFluentTagHost struct {
						Logs struct {
							Include struct {
								MatchType        *string `yaml:"match_type,omitempty"`
								RecordAttributes []struct {
									Key   *string `yaml:"key,omitempty"`
									Value *string `yaml:"value,omitempty"`
								} `yaml:"record_attributes,omitempty"`
							} `yaml:"include,omitempty"`
						} `yaml:"logs,omitempty"`
					} `yaml:"filter/include_fluent_tag_host,omitempty"`
					AttributesExtractSystemdSourceNameFromFluentTag struct {
						Actions []struct {
							Action  *string `yaml:"action,omitempty"`
							Key     *string `yaml:"key,omitempty"`
							Pattern *string `yaml:"pattern,omitempty"`
						} `yaml:"actions,omitempty"`
					} `yaml:"attributes/extract_systemd_source_name_from_fluent_tag,omitempty"`
					FilterIncludeSystemd struct {
						Logs struct {
							Include struct {
								MatchType        *string `yaml:"match_type,omitempty"`
								RecordAttributes []struct {
									Key   *string `yaml:"key,omitempty"`
									Value *string `yaml:"value,omitempty"`
								} `yaml:"record_attributes,omitempty"`
							} `yaml:"include,omitempty"`
						} `yaml:"logs,omitempty"`
					} `yaml:"filter/include_systemd,omitempty"`
					FilterExcludeKubelet struct {
						Logs struct {
							Exclude struct {
								MatchType        *string `yaml:"match_type,omitempty"`
								RecordAttributes []struct {
									Key   *string `yaml:"key,omitempty"`
									Value *string `yaml:"value,omitempty"`
								} `yaml:"record_attributes,omitempty"`
							} `yaml:"exclude,omitempty"`
						} `yaml:"logs,omitempty"`
					} `yaml:"filter/exclude_kubelet,omitempty"`
					GroupbyattrsSystemd struct {
						Keys []*string `yaml:"keys,omitempty"`
					} `yaml:"groupbyattrs/systemd,omitempty"`
					SourceSystemd struct {
						Collector                 *string `yaml:"collector,omitempty"`
						SourceHost                *string `yaml:"source_host,omitempty"`
						SourceName                *string `yaml:"source_name,omitempty"`
						SourceCategory            *string `yaml:"source_category,omitempty"`
						SourceCategoryPrefix      *string `yaml:"source_category_prefix,omitempty"`
						SourceCategoryReplaceDash *string `yaml:"source_category_replace_dash,omitempty"`
						Exclude                   struct {
							SYSLOGFACILITY *string `yaml:"SYSLOG_FACILITY,omitempty"`
							HOSTNAME       *string `yaml:"_HOSTNAME,omitempty"`
							PRIORITY       *string `yaml:"PRIORITY,omitempty"`
							SYSTEMDUNIT    *string `yaml:"_SYSTEMD_UNIT,omitempty"`
						} `yaml:"exclude,omitempty"`
					} `yaml:"source/systemd,omitempty"`
					FilterIncludeKubelet struct {
						Logs struct {
							Include struct {
								MatchType        *string `yaml:"match_type,omitempty"`
								RecordAttributes []struct {
									Key   *string `yaml:"key,omitempty"`
									Value *string `yaml:"value,omitempty"`
								} `yaml:"record_attributes,omitempty"`
							} `yaml:"include,omitempty"`
						} `yaml:"logs,omitempty"`
					} `yaml:"filter/include_kubelet,omitempty"`
					SourceKubelet struct {
						Collector                 *string `yaml:"collector,omitempty"`
						SourceHost                *string `yaml:"source_host,omitempty"`
						SourceName                *string `yaml:"source_name,omitempty"`
						SourceCategory            *string `yaml:"source_category,omitempty"`
						SourceCategoryPrefix      *string `yaml:"source_category_prefix,omitempty"`
						SourceCategoryReplaceDash *string `yaml:"source_category_replace_dash,omitempty"`
						Exclude                   struct {
							SYSLOGFACILITY *string `yaml:"SYSLOG_FACILITY,omitempty"`
							HOSTNAME       *string `yaml:"_HOSTNAME,omitempty"`
							PRIORITY       *string `yaml:"PRIORITY,omitempty"`
							SYSTEMDUNIT    *string `yaml:"_SYSTEMD_UNIT,omitempty"`
						} `yaml:"exclude,omitempty"`
					} `yaml:"source/kubelet,omitempty"`
				} `yaml:"processors,omitempty"`
				Service struct {
					Telemetry struct {
						Logs struct {
							Level *string `yaml:"level,omitempty"`
						} `yaml:"logs,omitempty"`
					} `yaml:"telemetry,omitempty"`
					Extensions []*string `yaml:"extensions,omitempty"`
					Pipelines  struct {
						LogsFluentContainers struct {
							Receivers  []*string `yaml:"receivers,omitempty"`
							Processors []*string `yaml:"processors,omitempty"`
							Exporters  []*string `yaml:"exporters,omitempty"`
						} `yaml:"logs/fluent/containers,omitempty"`
						LogsFluentSystemd struct {
							Receivers  []*string `yaml:"receivers,omitempty"`
							Processors []*string `yaml:"processors,omitempty"`
							Exporters  []*string `yaml:"exporters,omitempty"`
						} `yaml:"logs/fluent/systemd,omitempty"`
						LogsFluentKubelet struct {
							Receivers  []*string `yaml:"receivers,omitempty"`
							Processors []*string `yaml:"processors,omitempty"`
							Exporters  []*string `yaml:"exporters,omitempty"`
						} `yaml:"logs/fluent/kubelet,omitempty"`
					} `yaml:"pipelines,omitempty"`
				} `yaml:"service,omitempty"`
			} `yaml:"config,omitempty"`
			Statefulset struct {
				NodeSelector    struct{}      `yaml:"nodeSelector,omitempty"`
				Tolerations     []interface{} `yaml:"tolerations,omitempty"`
				Affinity        struct{}      `yaml:"affinity,omitempty"`
				PodAntiAffinity *string       `yaml:"podAntiAffinity,omitempty"`
				ReplicaCount    *int          `yaml:"replicaCount,omitempty"`
				Resources       struct {
					Limits struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"limits,omitempty"`
					Requests struct {
						Memory *string `yaml:"memory,omitempty"`
						CPU    *string `yaml:"cpu,omitempty"`
					} `yaml:"requests,omitempty"`
				} `yaml:"resources,omitempty"`
				PriorityClassName interface{} `yaml:"priorityClassName,omitempty"`
				PodLabels         struct{}    `yaml:"podLabels,omitempty"`
				PodAnnotations    struct{}    `yaml:"podAnnotations,omitempty"`
				Containers        struct {
					Otelcol struct {
						SecurityContext struct{} `yaml:"securityContext,omitempty"`
						LivenessProbe   struct {
							InitialDelaySeconds *int `yaml:"initialDelaySeconds,omitempty"`
							PeriodSeconds       *int `yaml:"periodSeconds,omitempty"`
							TimeoutSeconds      *int `yaml:"timeoutSeconds,omitempty"`
							FailureThreshold    *int `yaml:"failureThreshold,omitempty"`
						} `yaml:"livenessProbe,omitempty"`
						ReadinessProbe struct {
							InitialDelaySeconds *int `yaml:"initialDelaySeconds,omitempty"`
							PeriodSeconds       *int `yaml:"periodSeconds,omitempty"`
							TimeoutSeconds      *int `yaml:"timeoutSeconds,omitempty"`
							FailureThreshold    *int `yaml:"failureThreshold,omitempty"`
						} `yaml:"readinessProbe,omitempty"`
						StartupProbe struct {
							PeriodSeconds    *int `yaml:"periodSeconds,omitempty"`
							FailureThreshold *int `yaml:"failureThreshold,omitempty"`
						} `yaml:"startupProbe,omitempty"`
					} `yaml:"otelcol,omitempty"`
				} `yaml:"containers,omitempty"`
			} `yaml:"statefulset,omitempty"`
			Autoscaling struct {
				Enabled                        *bool `yaml:"enabled,omitempty"`
				MinReplicas                    *int  `yaml:"minReplicas,omitempty"`
				MaxReplicas                    *int  `yaml:"maxReplicas,omitempty"`
				TargetCPUUtilizationPercentage *int  `yaml:"targetCPUUtilizationPercentage,omitempty"`
			} `yaml:"autoscaling,omitempty"`
			PodDisruptionBudget struct {
				MinAvailable *int `yaml:"minAvailable,omitempty"`
			} `yaml:"podDisruptionBudget,omitempty"`
		} `yaml:"logs,omitempty"`
	} `yaml:"metadata,omitempty"`
	Otellogs struct {
		Enabled *bool `yaml:"enabled,omitempty"`
		Image   struct {
			Repository *string `yaml:"repository,omitempty"`
			Tag        *string `yaml:"tag,omitempty"`
			PullPolicy *string `yaml:"pullPolicy,omitempty"`
		} `yaml:"image,omitempty"`
		LogLevel *string `yaml:"logLevel,omitempty"`
		Config   struct {
			Extensions struct {
				HealthCheck struct{} `yaml:"health_check,omitempty"`
				FileStorage struct {
					Directory *string `yaml:"directory,omitempty"`
					Timeout   *string `yaml:"timeout,omitempty"`
				} `yaml:"file_storage,omitempty"`
				Pprof struct{} `yaml:"pprof,omitempty"`
			} `yaml:"extensions,omitempty"`
			Service struct {
				Telemetry struct {
					Logs struct {
						Level *string `yaml:"level,omitempty"`
					} `yaml:"logs,omitempty"`
				} `yaml:"telemetry,omitempty"`
				Extensions []*string `yaml:"extensions,omitempty"`
				Pipelines  struct {
					LogsContainers struct {
						Receivers  []*string `yaml:"receivers,omitempty"`
						Processors []*string `yaml:"processors,omitempty"`
						Exporters  []*string `yaml:"exporters,omitempty"`
					} `yaml:"logs/containers,omitempty"`
				} `yaml:"pipelines,omitempty"`
			} `yaml:"service,omitempty"`
			Receivers struct {
				FilelogContainers struct {
					Include         []*string `yaml:"include,omitempty"`
					StartAt         *string   `yaml:"start_at,omitempty"`
					IncludeFilePath *bool     `yaml:"include_file_path,omitempty"`
					IncludeFileName *bool     `yaml:"include_file_name,omitempty"`
					Operators       []struct {
						ID     *string `yaml:"id,omitempty"`
						Type   *string `yaml:"type,omitempty"`
						Routes []struct {
							Output *string `yaml:"output,omitempty"`
							Expr   *string `yaml:"expr,omitempty"`
						} `yaml:"routes,omitempty"`
						Regex     *string `yaml:"regex,omitempty,omitempty"`
						Output    *string `yaml:"output,omitempty,omitempty"`
						Timestamp struct {
							ParseFrom  *string `yaml:"parse_from,omitempty"`
							LayoutType *string `yaml:"layout_type,omitempty"`
							Layout     *string `yaml:"layout,omitempty"`
						} `yaml:"timestamp,omitempty"`
						CombineField  *string `yaml:"combine_field,omitempty,omitempty"`
						IsLastEntry   *string `yaml:"is_last_entry,omitempty,omitempty"`
						CombineWith   *string `yaml:"combine_with,omitempty,omitempty"`
						OverwriteWith *string `yaml:"overwrite_with,omitempty,omitempty"`
						ParseFrom     *string `yaml:"parse_from,omitempty,omitempty"`
						Attributes    struct {
							Stream           *string `yaml:"stream,omitempty"`
							K8SContainerName *string `yaml:"k8s.container.name,omitempty"`
							K8SNamespaceName *string `yaml:"k8s.namespace.name,omitempty"`
							K8SPodName       *string `yaml:"k8s.pod.name,omitempty"`
							RunID            *string `yaml:"run_id,omitempty"`
							K8SPodUID        *string `yaml:"k8s.pod.uid,omitempty"`
						} `yaml:"attributes,omitempty"`
						Ops []struct {
							Move struct {
								From *string `yaml:"from,omitempty"`
								To   *string `yaml:"to,omitempty"`
							} `yaml:"move,omitempty"`
						} `yaml:"ops,omitempty"`
					} `yaml:"operators,omitempty"`
				} `yaml:"filelog/containers,omitempty"`
			} `yaml:"receivers,omitempty"`
			Exporters struct {
				Otlphttp struct {
					Endpoint *string `yaml:"endpoint,omitempty"`
				} `yaml:"otlphttp,omitempty"`
			} `yaml:"exporters,omitempty"`
			Processors struct {
				Batch struct {
					SendBatchSize *int    `yaml:"send_batch_size,omitempty"`
					Timeout       *string `yaml:"timeout,omitempty"`
				} `yaml:"batch,omitempty"`
			} `yaml:"processors,omitempty"`
		} `yaml:"config,omitempty"`
		Daemonset struct {
			SecurityContext struct {
				FsGroup    *int `yaml:"fsGroup,omitempty"`
				RunAsUser  *int `yaml:"runAsUser,omitempty"`
				RunAsGroup *int `yaml:"runAsGroup,omitempty"`
			} `yaml:"securityContext,omitempty"`
			PodLabels      struct{} `yaml:"podLabels,omitempty"`
			PodAnnotations struct{} `yaml:"podAnnotations,omitempty"`
			Resources      struct {
				Limits struct {
					Memory *string `yaml:"memory,omitempty"`
					CPU    *string `yaml:"cpu,omitempty"`
				} `yaml:"limits,omitempty"`
				Requests struct {
					Memory *string `yaml:"memory,omitempty"`
					CPU    *string `yaml:"cpu,omitempty"`
				} `yaml:"requests,omitempty"`
			} `yaml:"resources,omitempty"`
			PriorityClassName interface{} `yaml:"priorityClassName,omitempty"`
			Containers        struct {
				Otelcol struct {
					SecurityContext struct{} `yaml:"securityContext,omitempty"`
				} `yaml:"otelcol,omitempty"`
			} `yaml:"containers,omitempty"`
			NodeSelector struct{}      `yaml:"nodeSelector,omitempty"`
			Tolerations  []interface{} `yaml:"tolerations,omitempty"`
			Affinity     struct{}      `yaml:"affinity,omitempty"`
		} `yaml:"daemonset,omitempty"`
	} `yaml:"otellogs,omitempty"`
	TelegrafOperator struct {
		Enabled *bool `yaml:"enabled,omitempty"`
		Image   struct {
			SidecarImage *string `yaml:"sidecarImage,omitempty"`
		} `yaml:"image,omitempty"`
		ReplicaCount *int `yaml:"replicaCount,omitempty"`
		Classes      struct {
			SecretName *string `yaml:"secretName,omitempty"`
			Default    *string `yaml:"default,omitempty"`
			Data       struct {
				SumologicPrometheus *string `yaml:"sumologic-prometheus,omitempty"`
			} `yaml:"data,omitempty"`
		} `yaml:"classes,omitempty"`
	} `yaml:"telegraf-operator,omitempty"`
	Falco struct {
		Enabled *bool `yaml:"enabled,omitempty"`
		Image   struct {
			Registry   *string `yaml:"registry,omitempty"`
			Repository *string `yaml:"repository,omitempty"`
		} `yaml:"image,omitempty"`
		AddKernelDevel      *bool `yaml:"addKernelDevel,omitempty"`
		ExtraInitContainers []struct {
			Name         *string   `yaml:"name,omitempty"`
			Image        *string   `yaml:"image,omitempty"`
			Command      []*string `yaml:"command,omitempty"`
			VolumeMounts []struct {
				MountPath *string `yaml:"mountPath,omitempty"`
				Name      *string `yaml:"name,omitempty"`
				ReadOnly  *bool   `yaml:"readOnly,omitempty"`
			} `yaml:"volumeMounts,omitempty"`
		} `yaml:"extraInitContainers,omitempty"`
		Falco struct {
			JSONOutput *bool     `yaml:"jsonOutput,omitempty"`
			RulesFile  []*string `yaml:"rulesFile,omitempty"`
		} `yaml:"falco,omitempty"`
		CustomRules struct {
			RulesUserKnownK8SAPICallersYaml       *string `yaml:"rules_user_known_k8s_api_callers.yaml,omitempty"`
			RulesUserSensitiveMountContainersYaml *string `yaml:"rules_user_sensitive_mount_containers.yaml,omitempty"`
			RulesUserPrivilegedContainersYaml     *string `yaml:"rules_user_privileged_containers.yaml,omitempty"`
		} `yaml:"customRules,omitempty"`
	} `yaml:"falco,omitempty"`
	TailingSidecarOperator struct {
		Enabled *bool `yaml:"enabled,omitempty"`
	} `yaml:"tailing-sidecar-operator,omitempty"`
}
