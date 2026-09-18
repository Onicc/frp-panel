<template>
  <div class="topology-map-shell">
    <div class="topology-map-canvas">
      <VMap
        :key="mapKey"
        :options="mapOptions"
        projection="globe"
        class="topology-map"
        @loaded="onMapLoaded"
        @move="onMapMove"
      >
        <VControlAttribution :options="{ compact: true }" position="bottom-right" />
        <VControlNavigation position="top-right" />
        <VControlScale position="bottom-left" />

        <VMarker
          v-for="node in locatedNodes"
          :key="node.id"
          :coordinates="node.position"
          :options="{ anchor: 'center' }"
          @click="selectNode(node.id)"
        >
          <template #markers="{ setRef }">
            <button
              :ref="(value) => assignMarkerRef(setRef, value)"
              class="topology-marker"
              :class="[node.kind, { selected: selectedNodeId === node.id }]"
              type="button"
              :aria-label="node.label"
              @click.stop="selectNode(node.id)"
            >
              <Icon :name="node.kind === 'server' ? 'server' : 'users'" size="xs" />
            </button>
          </template>
        </VMarker>
      </VMap>

      <div v-if="!nodes.length" class="topology-map-empty">
        <Icon name="globe" size="lg" />
        <strong>{{ t('overview.topologyEmpty') }}</strong>
        <span>{{ t('overview.mapNoLocation') }}</span>
      </div>
      <div v-else-if="!locatedNodes.length" class="topology-map-empty">
        <Icon name="globe" size="lg" />
        <strong>{{ t('overview.mapNoLocation') }}</strong>
        <span>{{ t('overview.mapDataSource') }}</span>
      </div>
    </div>

    <aside class="topology-map-overlay" aria-label="Network topology details">
      <div class="topology-map-heading">
        <div>
          <span class="eyebrow"><Icon name="globe" size="xs" />{{ t('overview.mapMaster') }}</span>
          <strong>{{ t('overview.mapConnections') }}</strong>
        </div>
        <button class="topology-fit-button" type="button" :title="t('overview.mapFit')" @click="fitMap">
          <Icon name="globe" size="xs" />{{ t('overview.mapFit') }}
        </button>
      </div>

      <div class="topology-map-metrics">
        <span><strong>{{ locatedNodes.length }}</strong> {{ t('overview.mapLocated') }}</span>
        <span><strong>{{ nodes.length - locatedNodes.length }}</strong> {{ t('overview.mapUnlocated') }}</span>
        <span><strong>{{ links.length }}</strong> {{ t('overview.mapConnections') }}</span>
      </div>

      <div v-if="nodes.length" class="topology-node-list">
        <button
          v-for="node in nodes"
          :key="node.id"
          class="topology-node-row"
          :class="{ selected: selectedNodeId === node.id }"
          type="button"
          @click="focusNode(node)"
        >
          <span class="topology-node-kind" :class="node.kind"><Icon :name="node.kind === 'server' ? 'server' : 'users'" size="xs" /></span>
          <span class="topology-node-copy">
            <strong>{{ node.label }}</strong>
            <small>{{ geoLabel(node) }}</small>
          </span>
          <span class="topology-node-status" :class="node.status"><i></i>{{ t(`common.${node.status}`, node.status) }}</span>
        </button>
      </div>
      <p v-else class="topology-map-note">{{ t('overview.mapNoLocation') }}</p>

      <div class="topology-links">
        <div class="topology-section-label">{{ t('overview.mapConnections') }}</div>
        <div v-if="links.length" class="topology-link-list">
          <div v-for="link in links" :key="link.id" class="topology-link-row">
            <span class="topology-link-line" :class="link.status"></span>
            <span class="topology-link-copy"><strong>{{ shortId(link.sourceClientId) }} → {{ shortId(link.targetServerId) }}</strong><small>{{ link.name }} · {{ link.type.toUpperCase() }} :{{ link.remotePort }}</small></span>
            <span class="topology-node-status" :class="link.status"><i></i>{{ t(`common.${link.status}`, link.status) }}</span>
          </div>
        </div>
        <p v-else class="topology-map-note">{{ t('common.noData') }}</p>
      </div>

      <div v-if="selectedNode" class="topology-node-detail">
        <div class="topology-section-label">{{ selectedNode.kind === 'server' ? t('overview.mapServer') : t('overview.mapClient') }}</div>
        <strong>{{ selectedNode.label }}</strong>
        <small v-if="selectedNode.locationIp">{{ selectedNode.locationIp }} · {{ geoLabel(selectedNode) }}</small>
        <small v-else>{{ t('overview.mapUnlocated') }}</small>
        <small v-if="selectedNode.lastSeenAt">{{ t('overview.mapLastSeen') }}: {{ formatDate(selectedNode.lastSeenAt) }}</small>
      </div>

      <p class="topology-map-source"><Icon name="infoCircle" size="xs" />{{ t('overview.mapDataSource') }}</p>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { VControlAttribution, VControlNavigation, VControlScale, VMap, VMarker } from '@geoql/v-maplibre'
import { LngLatBounds, type ExpressionSpecification, type GeoJSONSource, type Map as MapLibreMap } from 'maplibre-gl'
import type { TopologyLink, TopologyNode } from '../../api'
import Icon from '../icons/Icon.vue'
import { useTheme } from '../../theme'
import { createIpGeoLookup, type IpGeoEntry } from '../../utils/ipGeoLookup'

type LocatedNode = TopologyNode & { position: [number, number]; geo?: IpGeoEntry }
type ArcDatum = { link: TopologyLink; source: [number, number]; target: [number, number] }

const props = defineProps<{ nodes: TopologyNode[]; links: TopologyLink[] }>()
const { t, locale } = useI18n()
const theme = useTheme()
const geo = createIpGeoLookup()
const map = shallowRef<MapLibreMap>()
const selectedNodeId = ref('')
const hasFitted = ref(false)
const viewState = ref({ center: [0, 18] as [number, number], zoom: 1.15 })
const mapId = `frp-topology-${Math.random().toString(36).slice(2)}`

const mapStyle = computed(() => theme.resolved.value === 'dark'
  ? 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json'
  : 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json')
const mapKey = computed(() => `${theme.resolved.value}-${mapStyle.value}`)
const mapOptions = computed(() => ({
  container: mapId,
  style: mapStyle.value,
  center: viewState.value.center,
  zoom: viewState.value.zoom,
  pitch: 0,
  bearing: 0,
}))

const geoEntry = (node: TopologyNode) => node.locationIp ? geo.entries.value[node.locationIp] : undefined
const locatedNodes = computed<LocatedNode[]>(() => props.nodes.flatMap((node) => {
  const detail = geoEntry(node)?.detail
  if (!detail || detail.latitude === undefined || detail.longitude === undefined) return []
  return [{ ...node, position: [detail.longitude, detail.latitude], geo: geoEntry(node) }]
}))
const positionById = computed(() => new Map(locatedNodes.value.map((node) => [node.id, node.position])))
const arcData = computed<ArcDatum[]>(() => props.links.flatMap((link) => {
  const source = positionById.value.get(link.sourceClientId)
  const target = positionById.value.get(link.targetServerId)
  return source && target ? [{ link, source, target }] : []
}))
const selectedNode = computed(() => props.nodes.find((node) => node.id === selectedNodeId.value))

watch(() => props.nodes.map((node) => node.locationIp || '').filter(Boolean).join('|'), (value) => {
  void geo.lookupMany(value ? value.split('|') : [])
}, { immediate: true })
watch(() => locatedNodes.value.map((node) => `${node.position[0]},${node.position[1]}`).join('|'), () => {
  if (map.value && !hasFitted.value) fitMap()
})
watch(() => theme.resolved.value, () => { hasFitted.value = false })

function assignMarkerRef(setRef: (element: Element | HTMLElement | null) => void, value: unknown) {
  setRef(value instanceof Element ? value : null)
}

function greatCirclePath(source: [number, number], target: [number, number], segments = 40): [number, number][] {
  const radians = Math.PI / 180
  const degrees = 180 / Math.PI
  const a = [Math.cos(source[1] * radians) * Math.cos(source[0] * radians), Math.cos(source[1] * radians) * Math.sin(source[0] * radians), Math.sin(source[1] * radians)]
  const b = [Math.cos(target[1] * radians) * Math.cos(target[0] * radians), Math.cos(target[1] * radians) * Math.sin(target[0] * radians), Math.sin(target[1] * radians)]
  const dot = Math.min(1, Math.max(-1, a[0] * b[0] + a[1] * b[1] + a[2] * b[2]))
  const angle = Math.acos(dot)
  if (angle < 0.0001) return [source, target]
  const sinAngle = Math.sin(angle)
  return Array.from({ length: segments + 1 }, (_, index) => {
    const progress = index / segments
    const scaleA = Math.sin((1 - progress) * angle) / sinAngle
    const scaleB = Math.sin(progress * angle) / sinAngle
    const x = scaleA * a[0] + scaleB * b[0]
    const y = scaleA * a[1] + scaleB * b[1]
    const z = scaleA * a[2] + scaleB * b[2]
    return [Math.atan2(y, x) * degrees, Math.atan2(z, Math.sqrt(x * x + y * y)) * degrees]
  })
}

const connectionGeoJSON = computed(() => ({
  type: 'FeatureCollection' as const,
  features: arcData.value.map((arc) => ({
    type: 'Feature' as const,
    geometry: { type: 'LineString' as const, coordinates: greatCirclePath(arc.source, arc.target) },
    properties: { status: arc.link.status },
  })),
}))

const connectionSourceId = 'frp-topology-connections'
const connectionGlowLayerId = 'frp-topology-connections-glow'
const connectionLayerId = 'frp-topology-connections'

function syncConnections() {
  if (!map.value) return
  const data = connectionGeoJSON.value
  const source = map.value.getSource(connectionSourceId) as GeoJSONSource | undefined
  if (source) {
    source.setData(data)
    return
  }
  map.value.addSource(connectionSourceId, { type: 'geojson', data })
  const colorExpression = ['match', ['get', 'status'], 'online', '#2dd4bf', 'error', '#f87171', 'pending', '#f59e0b', '#94a3b8'] as unknown as ExpressionSpecification
  map.value.addLayer({ id: connectionGlowLayerId, type: 'line', source: connectionSourceId, paint: { 'line-color': colorExpression, 'line-width': 8, 'line-opacity': 0.18, 'line-blur': 4 } })
  map.value.addLayer({ id: connectionLayerId, type: 'line', source: connectionSourceId, paint: { 'line-color': colorExpression, 'line-width': 1.8, 'line-opacity': 0.9, 'line-dasharray': [2, 1.5] } })
}

watch(connectionGeoJSON, syncConnections, { deep: true })

function onMapLoaded(value: unknown) {
  map.value = value as MapLibreMap
  syncConnections()
  if (!hasFitted.value) fitMap()
}
function onMapMove(event: { target?: MapLibreMap }) {
  const target = event?.target
  if (!target) return
  const center = target.getCenter()
  viewState.value = { center: [center.lng, center.lat], zoom: target.getZoom() }
}
function fitMap() {
  if (!map.value || !locatedNodes.value.length) return
  if (locatedNodes.value.length === 1) {
    map.value.easeTo({ center: locatedNodes.value[0].position, zoom: 2.4, duration: 450 })
  } else {
    const bounds = new LngLatBounds()
    locatedNodes.value.forEach((node) => bounds.extend(node.position))
    map.value.fitBounds(bounds, { padding: 70, maxZoom: 3.2, duration: 550 })
  }
  hasFitted.value = true
}
function selectNode(id: string) { selectedNodeId.value = id }
function focusNode(node: TopologyNode) {
  selectedNodeId.value = node.id
  const position = positionById.value.get(node.id)
  if (position && map.value) map.value.easeTo({ center: position, zoom: Math.max(map.value.getZoom(), 2.2), duration: 450 })
}
function shortId(value: string) { return value.split('.').pop()?.slice(0, 12) || value.slice(0, 12) }
function geoLabel(node: TopologyNode) {
  const entry = geoEntry(node)
  if (entry?.status === 'loading') return t('overview.mapLoading')
  if (entry?.status === 'private') return t('overview.mapPrivate')
  if (entry?.status === 'error') return t('overview.mapGeoError')
  const detail = entry?.detail
  return [detail?.countryCode, detail?.region, detail?.city].filter(Boolean).join(' · ') || t('overview.mapUnlocated')
}
function formatDate(value: string) {
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

onBeforeUnmount(() => { map.value = undefined })
</script>
