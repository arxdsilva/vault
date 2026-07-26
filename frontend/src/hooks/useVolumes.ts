import { useEffect, useState } from 'react'
import { vault, type Volume } from '../lib/vault'

// Mounted external volumes (USB drives, disk images, network shares),
// alongside pinned locations in the side panel (spec §4.3). The Go side
// polls for mount/unmount changes and emits an event; refetch on it so a
// plugged-in drive shows up without user action.
export function useVolumes() {
  const [volumes, setVolumes] = useState<Volume[]>([])

  function refresh() {
    return vault
      .listVolumes()
      .then(setVolumes)
      .catch(() => {})
  }

  useEffect(() => {
    refresh()
    return vault.onVolumesChange(refresh)
  }, [])

  // Optimistically drop the volume so the rail feels instant; if the eject
  // actually failed (e.g. the volume is busy), re-fetch to put it back.
  function eject(path: string) {
    setVolumes((prev) => prev.filter((v) => v.path !== path))
    return vault.ejectVolume(path).catch(() => {
      refresh()
    })
  }

  return { volumes, eject }
}
