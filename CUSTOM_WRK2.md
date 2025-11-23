# wrk2 with Native Continuous Mode

This repository includes a modified version of wrk2 that supports native continuous/infinite mode.

## Modification Details

The wrk2 source code in this repository has been modified to support infinite duration when `duration` is set to `0`.

### Changes Made to wrk2

In `wrk2/src/wrk.c`, the duration handling has been modified:

```c
// Original code (line ~169):
uint64_t stop_at = time_us() + (cfg.duration * 1000000);

// Modified to:
uint64_t stop_at;
if (cfg.duration == 0) {
    // Duration is 0, run infinitely
    stop_at = UINT64_MAX;
} else {
    stop_at = time_us() + (cfg.duration * 1000000);
}
```

This allows wrk2 to run infinitely when duration is set to 0, providing native continuous load testing without the need for shell loop wrappers.

## Usage

The load generator Dockerfiles automatically build this modified wrk2. No additional setup is required.

### Setting Infinite Duration

In your Helm values.yaml:

```yaml
loadTest:
  duration: "0"  # Run infinitely
```

Or for a finite duration:

```yaml
loadTest:
  duration: "5m"  # Run for 5 minutes
```

## Building

The modified wrk2 is built automatically as part of the Dockerfile-loader build process:

```bash
cd hotelReservation  # or socialNetwork or mediaMicroservices
docker build -f Dockerfile-loader -t registry/service-loader:latest .
```

## Technical Details

- When `duration=0`, `stop_at` is set to `UINT64_MAX` (maximum uint64 value)
- The load test will run until manually stopped or the pod is terminated
- All wrk2 functionality remains unchanged except for the infinite duration support
- The modification is minimal and only affects duration handling

