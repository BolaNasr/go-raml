"""
Auto-generated class for Petshop
"""
import capnp
import os
from .Cat import Cat
from typing import List

from . import client_support

dir = os.path.dirname(os.path.realpath(__file__))


class Petshop:
    """
    auto-generated. don't touch.
    """

    def __init__(self, cats: List[Cat], name: str) -> None:
        """
        :type cats: list[Cat]
        :type name: str
        :rtype: Petshop
        """
        self.cats = cats  # type: List[Cat]
        self.name = name  # type: str

    def to_capnp(self):
        """
        Load the class in capnp schema Petshop.capnp
        :rtype bytes
        """
        template = capnp.load('%s/Petshop.capnp' % dir)
        return template.Petshop.new_message(**self.as_dict()).to_bytes()

    def as_dict(self):
        return client_support.to_dict(self)


class PetshopCollection:
    """
    auto-generated. don't touch.
    """

    @staticmethod
    def new(bin=None) -> Petshop:
        """
        Load the binary of Petshop.capnp into class Petshop
        :type bin: bytes. If none creates an empty capnp object.
        rtype: Petshop
        """
        template = capnp.load('%s/Petshop.capnp' % dir)
        struct = template.Petshop.from_bytes(bin) if bin else template.Petshop.new_message()
        return Petshop(**struct.to_dict(verbose=True))