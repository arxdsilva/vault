import * as Files from '../../bindings/github.com/arxdsilva/vault/internal/service/files'
import * as SearchService from '../../bindings/github.com/arxdsilva/vault/internal/service/search'
import * as PinnedService from '../../bindings/github.com/arxdsilva/vault/internal/service/pinned'
import * as VolumesService from '../../bindings/github.com/arxdsilva/vault/internal/service/volumes'
import * as PermissionsService from '../../bindings/github.com/arxdsilva/vault/internal/service/permissions'
import { Clipboard, Events } from '@wailsio/runtime'

export interface Entry {
  name: string
  path: string
  isDir: boolean
  size: number
  modTime: string
}

export interface IndexEntry {
  name: string
  path: string
  isDir: boolean
}

export interface RecentEntry {
  name: string
  path: string
  isDir: boolean
  lastAccess: string
}

export interface FsChange {
  dir: string
  path: string
  op: 'created' | 'removed' | 'modified' | 'renamed'
}

export interface Volume {
  path: string
  name: string
  ejectable: boolean
  totalBytes: number
  freeBytes: number
  isNetwork: boolean
}

export interface PermissionStatus {
  granted: boolean
  title: string
  message: string
  settingsLabel: string
}

const FS_CHANGE_EVENT = 'vault:fschange'
const VOLUMES_CHANGE_EVENT = 'vault:volumeschange'

// The Wails TS generator types every Go slice return as T[] | null, since a
// Go slice can always be nil — even where our own service methods guarantee
// a non-nil (possibly empty) slice at runtime. Normalize once here so the
// rest of the app can treat these as plain arrays.
function orEmpty<T>(v: T[] | null): T[] {
  return v ?? []
}

export const vault = {
  list: (dir: string) => Files.List(dir).then((r) => orEmpty(r) as unknown as Entry[]),
  stat: (path: string) => Files.Stat(path) as unknown as Promise<Entry>,
  homeDir: () => Files.HomeDir(),
  rename: (oldPath: string, newPath: string) => Files.Rename(oldPath, newPath),
  move: (src: string, dst: string) => Files.Move(src, dst),
  copy: (src: string, dst: string) => Files.Copy(src, dst),
  newFolder: (parent: string, name: string) => Files.NewFolder(parent, name),
  delete: (path: string) => Files.Delete(path),
  duplicate: (path: string) => Files.Duplicate(path),
  compress: (path: string) => Files.Compress(path),
  paste: (src: string, destDir: string) => Files.Paste(src, destDir),
  copyPathToClipboard: (path: string) => Clipboard.SetText(path),
  undo: () => Files.Undo(),
  trashLocation: () => Files.TrashLocation(),
  thumbnail: (path: string) => Files.Thumbnail(path),
  watch: (dir: string) => Files.Watch(dir),
  unwatch: () => Files.Unwatch(),

  onChange(handler: (change: FsChange) => void): () => void {
    return Events.On(FS_CHANGE_EVENT, (event) => {
      handler(event.data as FsChange)
    })
  },

  search: (query: string, limit: number, dir = '') =>
    SearchService.Query(query, limit, dir).then((r) => orEmpty(r) as unknown as IndexEntry[]),
  visit: (dir: string) => SearchService.Visit(dir),
  touch: (path: string) => SearchService.Touch(path),
  recent: (limit: number) =>
    SearchService.Recent(limit).then((r) => orEmpty(r) as unknown as RecentEntry[]),

  listPinned: () => PinnedService.List().then(orEmpty),
  addPinned: (path: string) => PinnedService.Add(path),
  removePinned: (path: string) => PinnedService.Remove(path),
  reorderPinned: (paths: string[]) => PinnedService.Reorder(paths),

  listVolumes: () => VolumesService.List().then((r) => orEmpty(r) as unknown as Volume[]),
  ejectVolume: (path: string) => VolumesService.Eject(path),

  onVolumesChange(handler: () => void): () => void {
    return Events.On(VOLUMES_CHANGE_EVENT, () => handler())
  },

  permissionStatus: () => PermissionsService.Status() as unknown as Promise<PermissionStatus>,
  openPermissionSettings: () => PermissionsService.OpenSettings(),
}
