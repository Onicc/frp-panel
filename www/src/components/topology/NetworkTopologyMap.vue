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
            <small>{{ geoLabel(node) }} · {{ locationSourceLabel(node) }}</small>
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
        <small v-if="selectedNode.locationIp">{{ locationSourceLabel(selectedNode) }}</small>
        <small v-if="selectedNode.observedIp && selectedNode.locationSource !== 'observed'">{{ t('overview.mapObservedIp') }}: {{ selectedNode.observedIp }}</small>
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
import { ArcLayer } from '@deck.gl/layers'
import { MapboxOverlay } from '@deck.gl/mapbox'
import { LngLatBounds, type Map as MapLibreMap } from 'maplibre-gl'
import type { TopologyLink, TopologyNode } from '../../api'
import Icon from '../icons/Icon.vue'
import { routeParallelArcs } from './arcGeometry'
import { useTheme } from '../../theme'
import { createIpGeoLookup, type IpGeoEntry } from '../../utils/ipGeoLookup'

type LocatedNode = TopologyNode & { position: [number, number]; geo?: IpGeoEntry }
type ArcColor = [number, number, number]
type ArcDatum = {
  link: TopologyLink
  source: [number, number]
  target: [number, number]
  sourceId: string
  targetId: string
  lane: number
  routeCount: number
  height: number
}

const props = defineProps<{ nodes: TopologyNode[]; links: TopologyLink[] }>()
const { t, locale } = useI18n()
const theme = useTheme()
const geo = createIpGeoLookup()
const map = shallowRef<MapLibreMap>()
const deckOverlay = shallowRef<MapboxOverlay>()
const selectedNodeId = ref('')
const hasFitted = ref(false)
const viewState = ref({ center: [0, 18] as [number, number], zoom: 1.15 })
const mapId = `frp-topology-${Math.random().toString(36).slice(2)}`

const mapStyle = computed(() => theme.resolved.value === 'dark'
  ? 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json'
  : 'https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json')
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
const arcCandidates = computed(() => props.links.flatMap((link) => {
  const source = positionById.value.get(link.sourceClientId)
  const target = positionById.value.get(link.targetServerId)
  return source && target ? [{ link, source, target, sourceId: link.sourceClientId, targetId: link.targetServerId }] : []
}))
const arcData = computed<ArcDatum[]>(() => routeParallelArcs(arcCandidates.value))
const selectedNode = computed(() => props.nodes.find((node) => node.id === selectedNodeId.value))
const locationSourceLabel = (node: TopologyNode) => {
  if (!node.locationSource) return t('overview.mapUnlocated')
  return t(`overview.locationSources.${node.locationSource}`)
}

watch(() => props.nodes.map((node) => node.locationIp || '').filter(Boolean).join('|'), (value) => {
  void geo.lookupMany(value ? value.split('|') : [])
}, { immediate: true })
watch(() => theme.resolved.value, () => { hasFitted.value = false })

function assignMarkerRef(setRef: (element: Element | HTMLElement | null) => void, value: unknown) {
  setRef(value instanceof Element ? value : null)
}

function arcColors(status: string): { source: ArcColor; target: ArcColor } {
  if (status === 'error') return { source: [248, 113, 113], target: [220, 38, 38] }
  if (status === 'pending') return { source: [251, 191, 36], target: [245, 158, 11] }
  if (status === 'online') return { source: [56, 189, 248], target: [45, 212, 191] }
  return { source: [148, 163, 184], target: [100, 116, 139] }
}

const arcSource = (arc: ArcDatum) => arc.source
const arcTarget = (arc: ArcDatum) => arc.target
const arcHeight = (arc: ArcDatum) => arc.height
const arcWidth = (arc: ArcDatum) => arc.link.status === 'online' ? 2.2 : 1.8
const arcGlowWidth = (arc: ArcDatum) => arc.link.status === 'online' ? 8 : 6
const arcSourceColor = (arc: ArcDatum) => arcColors(arc.link.status).source
const arcTargetColor = (arc: ArcDatum) => arcColors(arc.link.status).target

function syncDeckLayers() {
  if (!deckOverlay.value) return
  const data = arcData.value
  deckOverlay.value.setProps({ layers: data.length ? [
    new ArcLayer<ArcDatum>({
      id: 'frp-topology-arcs-glow',
      data,
      getSourcePosition: arcSource,
      getTargetPosition: arcTarget,
      getSourceColor: arcSourceColor,
      getTargetColor: arcTargetColor,
      getHeight: arcHeight,
      getWidth: arcGlowWidth,
      greatCircle: true,
      numSegments: 50,
      widthUnits: 'pixels',
      opacity: 0.18,
      pickable: false,
    }),
    new ArcLayer<ArcDatum>({
      id: 'frp-topology-arcs',
      data,
      getSourcePosition: arcSource,
      getTargetPosition: arcTarget,
      getSourceColor: arcSourceColor,
      getTargetColor: arcTargetColor,
      getHeight: arcHeight,
      getWidth: arcWidth,
      greatCircle: true,
      numSegments: 50,
      widthUnits: 'pixels',
      opacity: 0.92,
      pickable: false,
    }),
  ] : [] })
}

watch(arcData, syncDeckLayers, { deep: true })

function destroyDeckOverlay() {
  if (!map.value || !deckOverlay.value) return
  map.value.removeControl(deckOverlay.value)
  deckOverlay.value.finalize()
  deckOverlay.value = undefined
}

function onMapLoaded(value: unknown) {
  destroyDeckOverlay()
  map.value = value as MapLibreMap
  deckOverlay.value = new MapboxOverlay({ interleaved: false, useDevicePixels: 1 })
  map.value.addControl(deckOverlay.value)
  syncDeckLayers()
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

onBeforeUnmount(() => {
  destroyDeckOverlay()
  map.value = undefined
})
</script>
