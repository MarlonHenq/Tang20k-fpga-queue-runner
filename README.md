# Tang20k-fpga-queue-runner
A minimal Go-based API and worker system that queues Verilog code submissions, synthesizes them using Yosys, and deploys the resulting bitstream to a Tang Primer 20K FPGA using OpenFPGALoader.


# Notas de Desenvolvimento

Para rodar o Redis no Docker use:
```bash
sudo docker run -p 6300:6379 redis
```

6300 é a porta que o Go vai usar para se conectar ao Redis (ou seja, para onde o docker está roteando no host).

Para rodar o Go API, use:
```bash
go run api.go
```