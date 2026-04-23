# DeathStarBench

Open-source benchmark suite for cloud microservices. DeathStarBench includes five end-to-end services, four for cloud systems, and one for cloud-edge systems running on drone swarms. 

## End-to-end Services <img src="microservices_bundle4.png" alt="suite-icon" width="40"/>

* Social Network (released)
* Media Service (released)
* Hotel Reservation (released)
* E-commerce site (in progress)
* Banking System (in progress)
* Drone coordination system (in progress)

## Building the wrk2 loader images

Each benchmark ships a `Dockerfile-loader` that builds a wrk2-based load
generator. These Dockerfiles `COPY ../wrk2 ...` from the repo-root
`wrk2/` tree (which contains the `wrk2/deps/luajit` git submodule), so
they MUST be built with the repo root as the docker build context:

```bash
git submodule update --init --recursive
docker build -f hotelReservation/Dockerfile-loader -t hotelreservation-loader:latest .
docker build -f socialNetwork/Dockerfile-loader    -t socialnetwork-loader:latest    .
docker build -f mediaMicroservices/Dockerfile-loader -t mediamicroservices-loader:latest .
```

A helper that does submodule init + all three builds is provided at
`scripts/build-loaders.sh`.

## License & Copyright 

DeathStarBench is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, version 2.

DeathStarBench is being developed by the [SAIL group](http://sail.ece.cornell.edu/) at Cornell University. 

## Publications

More details on the applications and a characterization of their behavior can be found at ["An Open-Source Benchmark Suite for Microservices and Their Hardware-Software Implications for Cloud and Edge Systems"](http://www.csl.cornell.edu/~delimitrou/papers/2019.asplos.microservices.pdf), Y. Gan et al., ASPLOS 2019. 

If you use this benchmark suite in your work, we ask that you please cite the paper above. 


## Beta-testing

If you are interested in joining the beta-testing group for DeathStarBench, send us an email at: <microservices-bench-L@list.cornell.edu>
