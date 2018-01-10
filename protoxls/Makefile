INC_PATH = ../include/libxl ../protobuf-3.4.1/src
LIB_PATH = ../libs
BIN_PATH = ../bin
LIB = xl protobuf pthread

CC = g++
CC_FLAG = 
LINK_FLAG = 

PRG = protoxls
OBJ = BinExporter.o ConfigStore.o ExcelParser.o LuaExporter.o ParseHelper.o ProtoExcel.o strconv.o option.pb.o main.o

all:$(PRG)

$(PRG):$(OBJ)
	$(CC) -o $@ $(OBJ) $(LINK_FLAG) $(LIB_PATH:%=-L%) $(LIB:%=-l%)

.SUFFIXES: .c .cc .o .cpp
.cpp.o:
	$(CC) $(CC_FLAG) $(INC_PATH:%=-I%) -c $*.cpp -o $*.o

.cc.o:
	$(CC) $(CC_FLAG) $(INC_PATH:%=-I%) -c $*.cc -o $*.o

.PRONY:clean
clean:
	rm -f $(OBJ) $(PRG)

install:
	cp -f $(PRG) $(BIN_PATH)
	cp -f $(LIB_PATH)/libxl.so $(BIN_PATH)
